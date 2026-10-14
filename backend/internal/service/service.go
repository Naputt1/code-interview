// Package service implements the manufacturing tracker's business rules on top
// of GORM. GORM acts as the repository layer here, so this package owns both
// the rules and the queries.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/naputt/code-interview/backend/internal/domain"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ---- Resources ---------------------------------------------------------

type ResourceInput struct {
	Name     string   `json:"name" binding:"required"`
	Kind     string   `json:"kind" binding:"required"`
	Unit     string   `json:"unit" binding:"required"`
	Capacity *float64 `json:"capacity"`
}

func validateResourceInput(name string, in ResourceInput) error {
	if name == "" {
		return domain.Validation("resource name is required")
	}
	switch in.Kind {
	case domain.KindMaterial, domain.KindMachine, domain.KindLabour:
	default:
		return domain.Validation("resource kind must be material, machine or labour")
	}
	if strings.TrimSpace(in.Unit) == "" {
		return domain.Validation("resource unit is required")
	}
	if in.Capacity != nil && *in.Capacity < 0 {
		return domain.Validation("resource capacity cannot be negative")
	}
	return nil
}

func (s *Service) ListResources(ctx context.Context) ([]domain.Resource, error) {
	resources := []domain.Resource{}
	if err := s.db.WithContext(ctx).Order("name").Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("list resources: %w", err)
	}
	return resources, nil
}

func (s *Service) CreateResource(ctx context.Context, in ResourceInput) (*domain.Resource, error) {
	name := strings.TrimSpace(in.Name)
	if err := validateResourceInput(name, in); err != nil {
		return nil, err
	}
	resource := &domain.Resource{
		Name:     name,
		Kind:     in.Kind,
		Unit:     strings.TrimSpace(in.Unit),
		Capacity: in.Capacity,
	}
	if err := s.db.WithContext(ctx).Create(resource).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, domain.Conflict("resource_exists", fmt.Sprintf("a resource named %q already exists", name))
		}
		return nil, fmt.Errorf("create resource: %w", err)
	}
	return resource, nil
}

func (s *Service) UpdateResource(ctx context.Context, id int64, in ResourceInput) (*domain.Resource, error) {
	name := strings.TrimSpace(in.Name)
	if err := validateResourceInput(name, in); err != nil {
		return nil, err
	}
	var resource domain.Resource
	if err := s.db.WithContext(ctx).First(&resource, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.NotFound("resource", id)
		}
		return nil, fmt.Errorf("load resource: %w", err)
	}
	resource.Name = name
	resource.Kind = in.Kind
	resource.Unit = strings.TrimSpace(in.Unit)
	resource.Capacity = in.Capacity
	if err := s.db.WithContext(ctx).Save(&resource).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, domain.Conflict("resource_exists", fmt.Sprintf("a resource named %q already exists", name))
		}
		return nil, fmt.Errorf("update resource: %w", err)
	}
	return &resource, nil
}

func (s *Service) DeleteResource(ctx context.Context, id int64) error {
	var resource domain.Resource
	if err := s.db.WithContext(ctx).First(&resource, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.NotFound("resource", id)
		}
		return fmt.Errorf("load resource: %w", err)
	}

	var usedByProcesses int64
	if err := s.db.WithContext(ctx).Model(&domain.ProcessResource{}).
		Where("resource_id = ?", id).Count(&usedByProcesses).Error; err != nil {
		return fmt.Errorf("check process usage: %w", err)
	}
	var consumed int64
	if err := s.db.WithContext(ctx).Model(&domain.Consumption{}).
		Where("resource_id = ?", id).Count(&consumed).Error; err != nil {
		return fmt.Errorf("check consumption: %w", err)
	}
	if usedByProcesses > 0 || consumed > 0 {
		return domain.Conflict("resource_in_use", fmt.Sprintf(
			"%s is still used by %d process recipe(s) and %d consumption record(s)",
			resource.Name, usedByProcesses, consumed))
	}

	if err := s.db.WithContext(ctx).Delete(&resource).Error; err != nil {
		return fmt.Errorf("delete resource: %w", err)
	}
	return nil
}

// ---- Processes ---------------------------------------------------------

type ProcessResourceInput struct {
	ResourceID int64   `json:"resourceId" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required"`
}

type ProcessInput struct {
	Name     string                  `json:"name" binding:"required"`
	WIPLimit *int                    `json:"wipLimit"`
	Inputs   *[]ProcessResourceInput `json:"inputs"`
	Position *int                    `json:"position"`
}

func validateProcessInput(name string, in ProcessInput) error {
	if name == "" {
		return domain.Validation("process name is required")
	}
	if in.WIPLimit != nil && *in.WIPLimit < 1 {
		return domain.Validation("wip limit must be at least 1")
	}
	if in.Position != nil && *in.Position < 1 {
		return domain.Validation("position must be at least 1")
	}
	if in.Inputs != nil {
		seen := map[int64]bool{}
		for _, input := range *in.Inputs {
			if input.ResourceID <= 0 {
				return domain.Validation("each input needs a resource")
			}
			if input.Quantity <= 0 {
				return domain.Validation("each input needs a positive quantity")
			}
			if seen[input.ResourceID] {
				return domain.Validation("a resource can only appear once per process")
			}
			seen[input.ResourceID] = true
		}
	}
	return nil
}

func (s *Service) ListProcesses(ctx context.Context) ([]domain.Process, error) {
	processes := []domain.Process{}
	if err := s.db.WithContext(ctx).
		Preload("Inputs.Resource").
		Order("position, id").
		Find(&processes).Error; err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}
	return processes, nil
}

func (s *Service) GetProcess(ctx context.Context, id int64) (*domain.Process, error) {
	var process domain.Process
	if err := s.db.WithContext(ctx).
		Preload("Inputs.Resource").
		First(&process, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.NotFound("process", id)
		}
		return nil, fmt.Errorf("load process: %w", err)
	}
	return &process, nil
}

func (s *Service) CreateProcess(ctx context.Context, in ProcessInput) (*domain.Process, error) {
	name := strings.TrimSpace(in.Name)
	if err := validateProcessInput(name, in); err != nil {
		return nil, err
	}

	var created domain.Process
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		position := 0
		if in.Position != nil {
			position = *in.Position
		} else {
			var next int
			if err := tx.Model(&domain.Process{}).
				Select("COALESCE(MAX(position), 0) + 1").Scan(&next).Error; err != nil {
				return fmt.Errorf("next position: %w", err)
			}
			position = next
		}

		created = domain.Process{Name: name, Position: position, WIPLimit: in.WIPLimit}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}

		inputs, err := buildInputs(tx, created.ID, in.Inputs)
		if err != nil {
			return err
		}
		if len(inputs) > 0 {
			if err := tx.Create(&inputs).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.Conflict("process_exists", fmt.Sprintf("a process named %q already exists", name))
		}
		var derr *domain.Error
		if errors.As(err, &derr) {
			return nil, derr
		}
		return nil, fmt.Errorf("create process: %w", err)
	}
	return s.GetProcess(ctx, created.ID)
}

func (s *Service) UpdateProcess(ctx context.Context, id int64, in ProcessInput) (*domain.Process, error) {
	name := strings.TrimSpace(in.Name)
	if err := validateProcessInput(name, in); err != nil {
		return nil, err
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var process domain.Process
		if err := tx.First(&process, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.NotFound("process", id)
			}
			return err
		}

		updates := map[string]any{"name": name, "wip_limit": in.WIPLimit}
		if in.Position != nil {
			updates["position"] = *in.Position
		}
		if err := tx.Model(&process).Updates(updates).Error; err != nil {
			return err
		}

		if in.Inputs != nil {
			if err := tx.Where("process_id = ?", id).Delete(&domain.ProcessResource{}).Error; err != nil {
				return err
			}
			inputs, err := buildInputs(tx, id, in.Inputs)
			if err != nil {
				return err
			}
			if len(inputs) > 0 {
				if err := tx.Create(&inputs).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.Conflict("process_exists", fmt.Sprintf("a process named %q already exists", name))
		}
		var derr *domain.Error
		if errors.As(err, &derr) {
			return nil, derr
		}
		return nil, fmt.Errorf("update process: %w", err)
	}
	return s.GetProcess(ctx, id)
}

func (s *Service) DeleteProcess(ctx context.Context, id int64) error {
	var process domain.Process
	if err := s.db.WithContext(ctx).First(&process, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.NotFound("process", id)
		}
		return fmt.Errorf("load process: %w", err)
	}

	var usedBy int64
	if err := s.db.WithContext(ctx).Model(&domain.OrderStep{}).
		Where("process_id = ?", id).Count(&usedBy).Error; err != nil {
		return fmt.Errorf("check order usage: %w", err)
	}
	if usedBy > 0 {
		return domain.Conflict("process_in_use", fmt.Sprintf(
			"%s is referenced by %d order step(s) and cannot be deleted", process.Name, usedBy))
	}

	if err := s.db.WithContext(ctx).Delete(&process).Error; err != nil {
		return fmt.Errorf("delete process: %w", err)
	}
	return nil
}

func buildInputs(tx *gorm.DB, processID int64, in *[]ProcessResourceInput) ([]domain.ProcessResource, error) {
	if in == nil {
		return nil, nil
	}
	inputs := make([]domain.ProcessResource, 0, len(*in))
	for _, item := range *in {
		var resource domain.Resource
		if err := tx.First(&resource, item.ResourceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, domain.Validation(fmt.Sprintf("resource %d does not exist", item.ResourceID))
			}
			return nil, err
		}
		inputs = append(inputs, domain.ProcessResource{
			ProcessID:  processID,
			ResourceID: item.ResourceID,
			Quantity:   item.Quantity,
		})
	}
	return inputs, nil
}

// ---- helpers -----------------------------------------------------------

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
