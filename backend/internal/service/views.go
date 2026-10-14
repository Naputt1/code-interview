package service

import (
	"context"
	"fmt"
	"time"

	"github.com/naputt/code-interview/backend/internal/domain"
)

// ---- Board: orders currently in progress at each step -------------------

type BoardStep struct {
	StepID    int64      `json:"stepId"`
	OrderID   int64      `json:"orderId"`
	Reference string     `json:"reference"`
	Quantity  int        `json:"quantity"`
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"startedAt"`
	// Startable is meaningful on queued steps: it is true only when every
	// earlier step of the order is done and the process has a free WIP slot,
	// so the board can offer a Start button that will really be accepted.
	Startable bool `json:"startable"`
}

type BoardColumn struct {
	Process domain.Process `json:"process"`
	Running []BoardStep    `json:"running"`
	Queued  []BoardStep    `json:"queued"`
}

type Board struct {
	Columns []BoardColumn `json:"columns"`
}

func (s *Service) Board(ctx context.Context) (*Board, error) {
	processes := []domain.Process{}
	if err := s.db.WithContext(ctx).
		Preload("Inputs.Resource").
		Order("position, id").
		Find(&processes).Error; err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}

	board := &Board{Columns: make([]BoardColumn, 0, len(processes))}
	for _, process := range processes {
		column := BoardColumn{Process: process, Running: []BoardStep{}}

		running := []BoardStep{}
		if err := s.db.WithContext(ctx).Table("order_steps AS os").
			Select("os.id AS step_id, os.order_id, o.reference, o.quantity, os.status, os.started_at").
			Joins("JOIN orders AS o ON o.id = os.order_id").
			Where("os.process_id = ? AND os.status = ? AND o.status <> ?", process.ID, domain.StepRunning, domain.OrderCancelled).
			Order("os.started_at, os.id").
			Scan(&running).Error; err != nil {
			return nil, fmt.Errorf("list running steps: %w", err)
		}
		column.Running = running

		// Queued steps whose turn it actually is. An order appears only in the
		// column of its current step, so a column shows what is running there
		// plus what is next in line for it — never the steps further down an
		// order's route. Oldest order first, so the column reads as a queue.
		queued := []BoardStep{}
		if err := s.db.WithContext(ctx).Table("order_steps AS os").
			Select("os.id AS step_id, os.order_id, o.reference, o.quantity, os.status, os.started_at").
			Joins("JOIN orders AS o ON o.id = os.order_id").
			Where(`os.process_id = ? AND os.status = ? AND o.status <> ?
				AND NOT EXISTS (
					SELECT 1 FROM order_steps AS earlier
					WHERE earlier.order_id = os.order_id
					  AND earlier.position < os.position
					  AND earlier.status <> ?
				)`, process.ID, domain.StepQueued, domain.OrderCancelled, domain.StepDone).
			Order("o.created_at, os.id").
			Scan(&queued).Error; err != nil {
			return nil, fmt.Errorf("list queued steps: %w", err)
		}
		// Every row above is next in line, so only the WIP limit can hold it back.
		slotFree := process.WIPLimit == nil || len(column.Running) < *process.WIPLimit
		for i := range queued {
			queued[i].Startable = slotFree
		}
		column.Queued = queued

		board.Columns = append(board.Columns, column)
	}
	return board, nil
}

// ---- Overview: status of all orders and resource consumption ------------

type OrderCounts struct {
	Pending    int `json:"pending"`
	InProgress int `json:"inProgress"`
	Done       int `json:"done"`
	Cancelled  int `json:"cancelled"`
	Total      int `json:"total"`
}

type ProcessLoad struct {
	ProcessID int64  `json:"processId"`
	Name      string `json:"name"`
	Position  int    `json:"position"`
	Running   int    `json:"running"`
	Queued    int    `json:"queued"`
	Done      int    `json:"done"`
}

type ResourceUsage struct {
	Resource  domain.Resource `json:"resource"`
	Consumed  float64         `json:"consumed"`
	Remaining *float64        `json:"remaining"`
	UsedBy    int             `json:"usedBy"`
}

type Overview struct {
	Orders    OrderCounts     `json:"orders"`
	Processes []ProcessLoad   `json:"processes"`
	Resources []ResourceUsage `json:"resources"`
}

func (s *Service) Overview(ctx context.Context) (*Overview, error) {
	overview := &Overview{
		Processes: []ProcessLoad{},
		Resources: []ResourceUsage{},
	}

	type statusCount struct {
		Status string
		Count  int
	}

	var orderCounts []statusCount
	if err := s.db.WithContext(ctx).Model(&domain.Order{}).
		Select("status, count(*) AS count").Group("status").Scan(&orderCounts).Error; err != nil {
		return nil, fmt.Errorf("count orders: %w", err)
	}
	for _, item := range orderCounts {
		overview.Orders.Total += item.Count
		switch item.Status {
		case domain.OrderPending:
			overview.Orders.Pending = item.Count
		case domain.OrderInProgress:
			overview.Orders.InProgress = item.Count
		case domain.OrderDone:
			overview.Orders.Done = item.Count
		case domain.OrderCancelled:
			overview.Orders.Cancelled = item.Count
		}
	}

	processes := []domain.Process{}
	if err := s.db.WithContext(ctx).Order("position, id").Find(&processes).Error; err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}
	for _, process := range processes {
		load := ProcessLoad{ProcessID: process.ID, Name: process.Name, Position: process.Position}
		var stepCounts []statusCount
		if err := s.db.WithContext(ctx).Model(&domain.OrderStep{}).
			Select("status, count(*) AS count").
			Where("process_id = ?", process.ID).
			Group("status").Scan(&stepCounts).Error; err != nil {
			return nil, fmt.Errorf("count steps: %w", err)
		}
		for _, item := range stepCounts {
			switch item.Status {
			case domain.StepRunning:
				load.Running = item.Count
			case domain.StepQueued:
				load.Queued = item.Count
			case domain.StepDone:
				load.Done = item.Count
			}
		}
		overview.Processes = append(overview.Processes, load)
	}

	resources := []domain.Resource{}
	if err := s.db.WithContext(ctx).Order("name").Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("list resources: %w", err)
	}
	for _, resource := range resources {
		var agg struct {
			Consumed float64
			UsedBy   int
		}
		if err := s.db.WithContext(ctx).Model(&domain.Consumption{}).
			Select("COALESCE(SUM(quantity), 0) AS consumed, COUNT(DISTINCT order_step_id) AS used_by").
			Where("resource_id = ?", resource.ID).
			Scan(&agg).Error; err != nil {
			return nil, fmt.Errorf("aggregate consumption: %w", err)
		}
		usage := ResourceUsage{Resource: resource, Consumed: agg.Consumed, UsedBy: agg.UsedBy}
		if resource.Capacity != nil {
			remaining := *resource.Capacity - agg.Consumed
			usage.Remaining = &remaining
		}
		overview.Resources = append(overview.Resources, usage)
	}

	return overview, nil
}

// ---- Consumption ledger -------------------------------------------------

func (s *Service) ListConsumptions(ctx context.Context) ([]domain.Consumption, error) {
	entries := []domain.Consumption{}
	if err := s.db.WithContext(ctx).Order("occurred_at DESC, id DESC").Limit(200).Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("list consumptions: %w", err)
	}
	return entries, nil
}
