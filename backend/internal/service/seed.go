package service

import (
	"context"
	"fmt"

	"github.com/naputt/code-interview/backend/internal/domain"
)

// SeedIfEmpty populates a small demo manufacturing line the first time the
// database has no processes, so a fresh checkout has something to show.
func (s *Service) SeedIfEmpty(ctx context.Context) error {
	var existing int64
	if err := s.db.WithContext(ctx).Model(&domain.Process{}).Count(&existing).Error; err != nil {
		return fmt.Errorf("count processes: %w", err)
	}
	if existing > 0 {
		return nil
	}

	type resourceSeed struct {
		name     string
		kind     string
		unit     string
		capacity *float64
	}
	capacity := func(v float64) *float64 { return &v }

	resources := []resourceSeed{
		{"Aluminium Sheet", domain.KindMaterial, "kg", capacity(60)},
		{"Cutting Blade", domain.KindMachine, "hour", nil},
		{"Welding Wire", domain.KindMaterial, "kg", capacity(20)},
		{"Welding Station", domain.KindMachine, "hour", nil},
		{"Paint", domain.KindMaterial, "litre", capacity(15)},
		{"Paint Booth", domain.KindMachine, "hour", nil},
		{"Assembler", domain.KindLabour, "hour", nil},
	}

	ids := map[string]int64{}
	for _, seed := range resources {
		created, err := s.CreateResource(ctx, ResourceInput{
			Name: seed.name, Kind: seed.kind, Unit: seed.unit, Capacity: seed.capacity,
		})
		if err != nil {
			return fmt.Errorf("seed resource %s: %w", seed.name, err)
		}
		ids[seed.name] = created.ID
	}

	limit := func(v int) *int { return &v }
	processes := []struct {
		name   string
		limit  *int
		inputs []ProcessResourceInput
	}{
		{"Cut", limit(2), []ProcessResourceInput{
			{ResourceID: ids["Aluminium Sheet"], Quantity: 0.8},
			{ResourceID: ids["Cutting Blade"], Quantity: 0.25},
		}},
		{"Weld", limit(1), []ProcessResourceInput{
			{ResourceID: ids["Welding Wire"], Quantity: 0.15},
			{ResourceID: ids["Welding Station"], Quantity: 0.5},
		}},
		{"Paint", limit(1), []ProcessResourceInput{
			{ResourceID: ids["Paint"], Quantity: 0.2},
			{ResourceID: ids["Paint Booth"], Quantity: 0.75},
		}},
		{"Assemble", limit(3), []ProcessResourceInput{
			{ResourceID: ids["Assembler"], Quantity: 0.5},
		}},
	}

	for _, seed := range processes {
		inputs := seed.inputs
		if _, err := s.CreateProcess(ctx, ProcessInput{
			Name: seed.name, WIPLimit: seed.limit, Inputs: &inputs,
		}); err != nil {
			return fmt.Errorf("seed process %s: %w", seed.name, err)
		}
	}

	// A few orders spread across the sequence so every state is visible.
	orders := []struct {
		reference string
		quantity  int
		start     int // start this many steps, completing all but the last
	}{
		{"BRK-1001", 4, 1},
		{"BRK-1002", 2, 2},
		{"BRK-1003", 1, 0},
	}

	for _, seed := range orders {
		order, err := s.CreateOrder(ctx, OrderInput{Reference: seed.reference, Quantity: seed.quantity})
		if err != nil {
			return fmt.Errorf("seed order %s: %w", seed.reference, err)
		}
		for i := 0; i < seed.start; i++ {
			step := order.Steps[i]
			if _, err := s.StartStep(ctx, order.ID, step.ID); err != nil {
				return fmt.Errorf("seed start %s step %d: %w", seed.reference, i, err)
			}
			if i < seed.start-1 {
				if _, err := s.CompleteStep(ctx, order.ID, step.ID); err != nil {
					return fmt.Errorf("seed complete %s step %d: %w", seed.reference, i, err)
				}
			}
		}
	}
	return nil
}
