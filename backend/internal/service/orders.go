package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/naputt/code-interview/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderInput struct {
	Reference string `json:"reference" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
}

func (s *Service) ListOrders(ctx context.Context) ([]domain.Order, error) {
	orders := []domain.Order{}
	if err := s.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB { return db.Order("position, id") }).
		Preload("Steps.Process").
		Preload("Steps.Process.Inputs.Resource").
		Order("created_at DESC, id DESC").
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	return orders, nil
}

func (s *Service) GetOrder(ctx context.Context, id int64) (*domain.Order, error) {
	var order domain.Order
	if err := s.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB { return db.Order("position, id") }).
		Preload("Steps.Process").
		Preload("Steps.Process.Inputs.Resource").
		Preload("Steps.Consumptions").
		First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.NotFound("order", id)
		}
		return nil, fmt.Errorf("load order: %w", err)
	}
	return &order, nil
}

// CreateOrder snapshots the current process sequence into queued order steps.
// Editing the process list later therefore never rewrites an existing order.
func (s *Service) CreateOrder(ctx context.Context, in OrderInput) (*domain.Order, error) {
	reference := strings.TrimSpace(in.Reference)
	if reference == "" {
		return nil, domain.Validation("order reference is required")
	}
	if in.Quantity <= 0 {
		return nil, domain.Validation("order quantity must be greater than zero")
	}

	var created domain.Order
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var processes []domain.Process
		if err := tx.Order("position, id").Find(&processes).Error; err != nil {
			return err
		}
		if len(processes) == 0 {
			return domain.Validation("define at least one process before creating orders")
		}

		created = domain.Order{Reference: reference, Quantity: in.Quantity, Status: domain.OrderPending}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}

		steps := make([]domain.OrderStep, 0, len(processes))
		for _, process := range processes {
			steps = append(steps, domain.OrderStep{
				OrderID:   created.ID,
				ProcessID: process.ID,
				Position:  process.Position,
				Status:    domain.StepQueued,
			})
		}
		return tx.Create(&steps).Error
	})
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			return nil, derr
		}
		return nil, fmt.Errorf("create order: %w", err)
	}
	return s.GetOrder(ctx, created.ID)
}

// StartStep advances an order by starting one of its steps. It enforces the
// process rules: sequence order, the step's WIP limit, and finite resource
// capacity. Starting a step writes the resource consumption to the ledger.
func (s *Service) StartStep(ctx context.Context, orderID, stepID int64) (*domain.Order, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order domain.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, orderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.NotFound("order", orderID)
			}
			return err
		}
		switch order.Status {
		case domain.OrderCancelled:
			return domain.Conflict("order_cancelled", "this order is cancelled")
		case domain.OrderDone:
			return domain.Conflict("order_completed", "this order is already complete")
		}

		var step domain.OrderStep
		if err := tx.Where("id = ? AND order_id = ?", stepID, orderID).First(&step).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.NotFound("order step", stepID)
			}
			return err
		}
		switch step.Status {
		case domain.StepRunning:
			return domain.Conflict("step_running", "this step is already running")
		case domain.StepDone:
			return domain.Conflict("step_done", "this step is already complete")
		}

		var blocking int64
		if err := tx.Model(&domain.OrderStep{}).
			Where("order_id = ? AND position < ? AND status <> ?", orderID, step.Position, domain.StepDone).
			Count(&blocking).Error; err != nil {
			return err
		}
		if blocking > 0 {
			return domain.Conflict("out_of_sequence", "complete the earlier steps first")
		}

		// Lock the process row so two orders cannot both pass the WIP check below.
		var process domain.Process
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&process, step.ProcessID).Error; err != nil {
			return err
		}

		if process.WIPLimit != nil {
			var running int64
			if err := tx.Model(&domain.OrderStep{}).
				Where("process_id = ? AND status = ?", process.ID, domain.StepRunning).
				Count(&running).Error; err != nil {
				return err
			}
			if int(running) >= *process.WIPLimit {
				return domain.Conflict("wip_limit_reached", fmt.Sprintf(
					"%s can only run %d order(s) at a time", process.Name, *process.WIPLimit))
			}
		}

		var inputs []domain.ProcessResource
		if err := tx.Where("process_id = ?", process.ID).Find(&inputs).Error; err != nil {
			return err
		}
		// Lock the resources in a stable order so two starts that share a resource
		// serialise instead of both passing the capacity check, and cannot deadlock
		// each other while doing it.
		sort.Slice(inputs, func(i, j int) bool { return inputs[i].ResourceID < inputs[j].ResourceID })

		now := time.Now()
		for _, input := range inputs {
			quantity := input.Quantity * float64(order.Quantity)

			var resource domain.Resource
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&resource, input.ResourceID).Error; err != nil {
				return err
			}
			if resource.Capacity != nil {
				var consumed float64
				if err := tx.Model(&domain.Consumption{}).
					Where("resource_id = ?", resource.ID).
					Select("COALESCE(SUM(quantity), 0)").Scan(&consumed).Error; err != nil {
					return err
				}
				if consumed+quantity > *resource.Capacity {
					return domain.Conflict("insufficient_resource", fmt.Sprintf(
						"not enough %s: this step needs %.2f %s and only %.2f %s is left",
						resource.Name, quantity, resource.Unit,
						*resource.Capacity-consumed, resource.Unit))
				}
			}

			entry := domain.Consumption{
				OrderStepID: step.ID,
				ResourceID:  resource.ID,
				Quantity:    quantity,
				OccurredAt:  now,
			}
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&domain.OrderStep{}).Where("id = ?", step.ID).
			Updates(map[string]any{"status": domain.StepRunning, "started_at": now}).Error; err != nil {
			return err
		}
		return tx.Model(&domain.Order{}).Where("id = ?", orderID).
			Update("status", domain.OrderInProgress).Error
	})
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			return nil, derr
		}
		return nil, fmt.Errorf("start step: %w", err)
	}
	return s.GetOrder(ctx, orderID)
}

func (s *Service) CompleteStep(ctx context.Context, orderID, stepID int64) (*domain.Order, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order domain.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, orderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.NotFound("order", orderID)
			}
			return err
		}

		var step domain.OrderStep
		if err := tx.Where("id = ? AND order_id = ?", stepID, orderID).First(&step).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.NotFound("order step", stepID)
			}
			return err
		}
		if step.Status != domain.StepRunning {
			return domain.Conflict("step_not_running", "only a running step can be completed")
		}

		now := time.Now()
		if err := tx.Model(&domain.OrderStep{}).Where("id = ?", step.ID).
			Updates(map[string]any{"status": domain.StepDone, "finished_at": now}).Error; err != nil {
			return err
		}

		var remaining int64
		if err := tx.Model(&domain.OrderStep{}).
			Where("order_id = ? AND status <> ?", orderID, domain.StepDone).
			Count(&remaining).Error; err != nil {
			return err
		}
		if remaining == 0 {
			return tx.Model(&domain.Order{}).Where("id = ?", orderID).
				Update("status", domain.OrderDone).Error
		}
		return nil
	})
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			return nil, derr
		}
		return nil, fmt.Errorf("complete step: %w", err)
	}
	return s.GetOrder(ctx, orderID)
}

func (s *Service) CancelOrder(ctx context.Context, id int64) (*domain.Order, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order domain.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.NotFound("order", id)
			}
			return err
		}
		switch order.Status {
		case domain.OrderDone:
			return domain.Conflict("order_completed", "a completed order cannot be cancelled")
		case domain.OrderCancelled:
			return domain.Conflict("order_cancelled", "this order is already cancelled")
		case domain.OrderPending:
			// Nothing has been consumed yet, so a pending order cancels cleanly.
		default:
			// Once a step has run its consumption is committed and the order has
			// to be carried through rather than abandoned.
			return domain.Conflict("order_started", "an order cannot be cancelled once work has started")
		}

		return tx.Model(&domain.Order{}).Where("id = ?", id).
			Update("status", domain.OrderCancelled).Error
	})
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			return nil, derr
		}
		return nil, fmt.Errorf("cancel order: %w", err)
	}
	return s.GetOrder(ctx, id)
}
