package service

import (
	"context"
	"errors"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/naputt/code-interview/backend/internal/domain"
	"github.com/naputt/code-interview/backend/internal/store"
	"gorm.io/gorm"
)

var ctx = context.Background()

// newService returns a Service backed by an empty test database. Tests are
// skipped (not failed) when no database is reachable, so `go test ./...` still
// runs the pure-unit packages on a machine without Postgres.
func newService(t *testing.T) *Service {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://naputt@127.0.0.1:5432/code_interview_test?sslmode=disable"
	}
	db, err := store.Open(dsn)
	if err != nil {
		t.Skipf("no test database: %v", err)
	}
	if err := store.Migrate(ctx, db); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	for _, table := range []string{
		"consumptions", "order_steps", "orders", "process_resources", "processes", "resources",
	} {
		if err := db.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE").Error; err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}
	return New(db)
}

func intPtr(v int) *int           { return &v }
func floatPtr(v float64) *float64 { return &v }

func mustResource(t *testing.T, s *Service, name, kind, unit string, capacity *float64) *domain.Resource {
	t.Helper()
	r, err := s.CreateResource(ctx, ResourceInput{Name: name, Kind: kind, Unit: unit, Capacity: capacity})
	if err != nil {
		t.Fatalf("create resource %s: %v", name, err)
	}
	return r
}

func mustProcess(t *testing.T, s *Service, name string, wip *int, inputs ...ProcessResourceInput) *domain.Process {
	t.Helper()
	p, err := s.CreateProcess(ctx, ProcessInput{Name: name, WIPLimit: wip, Inputs: &inputs})
	if err != nil {
		t.Fatalf("create process %s: %v", name, err)
	}
	return p
}

func mustOrder(t *testing.T, s *Service, reference string, quantity int) *domain.Order {
	t.Helper()
	o, err := s.CreateOrder(ctx, OrderInput{Reference: reference, Quantity: quantity})
	if err != nil {
		t.Fatalf("create order %s: %v", reference, err)
	}
	return o
}

func wantsCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %q, got nil", code)
	}
	var domainErr *domain.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected a domain.Error with code %q, got %v", code, err)
	}
	if domainErr.Code != code {
		t.Fatalf("expected code %q, got %q (%v)", code, domainErr.Code, err)
	}
}

func stepID(t *testing.T, order *domain.Order, position int) int64 {
	t.Helper()
	for _, step := range order.Steps {
		if step.Position == position {
			return step.ID
		}
	}
	t.Fatalf("order has no step at position %d", position)
	return 0
}

func consumed(t *testing.T, s *Service, resourceID int64) float64 {
	t.Helper()
	overview, err := s.Overview(ctx)
	if err != nil {
		t.Fatalf("overview: %v", err)
	}
	for _, usage := range overview.Resources {
		if usage.Resource.ID == resourceID {
			return usage.Consumed
		}
	}
	t.Fatalf("resource %d missing from overview", resourceID)
	return 0
}

// An order may only work on the next step once every earlier step in the
// sequence is done.
func TestStepMustStartInSequence(t *testing.T) {
	s := newService(t)
	sheet := mustResource(t, s, "Sheet", domain.KindMaterial, "kg", nil)
	mustProcess(t, s, "Cut", nil, ProcessResourceInput{ResourceID: sheet.ID, Quantity: 1})
	mustProcess(t, s, "Weld", nil, ProcessResourceInput{ResourceID: sheet.ID, Quantity: 1})
	order := mustOrder(t, s, "SEQ-1", 1)

	_, err := s.StartStep(ctx, order.ID, stepID(t, order, 2))
	wantsCode(t, err, "out_of_sequence")

	if _, err := s.StartStep(ctx, order.ID, stepID(t, order, 1)); err != nil {
		t.Fatalf("start first step: %v", err)
	}
	if _, err := s.CompleteStep(ctx, order.ID, stepID(t, order, 1)); err != nil {
		t.Fatalf("complete first step: %v", err)
	}
	if _, err := s.StartStep(ctx, order.ID, stepID(t, order, 2)); err != nil {
		t.Fatalf("start second step after first is done: %v", err)
	}
}

// A process is a workstation with a limited number of slots; a second order
// cannot start there while the limit is reached.
func TestWIPLimitBlocksConcurrentStarts(t *testing.T) {
	s := newService(t)
	sheet := mustResource(t, s, "Sheet", domain.KindMaterial, "kg", nil)
	mustProcess(t, s, "Cut", intPtr(1), ProcessResourceInput{ResourceID: sheet.ID, Quantity: 1})
	first := mustOrder(t, s, "WIP-1", 1)
	second := mustOrder(t, s, "WIP-2", 1)

	if _, err := s.StartStep(ctx, first.ID, stepID(t, first, 1)); err != nil {
		t.Fatalf("start first order: %v", err)
	}
	_, err := s.StartStep(ctx, second.ID, stepID(t, second, 1))
	wantsCode(t, err, "wip_limit_reached")
}

// Starting a step consumes the recipe quantities; a step that would push a
// finite resource past its capacity is rejected before any consumption is
// written.
func TestCapacityBlocksStartAndRecordsConsumption(t *testing.T) {
	s := newService(t)
	sheet := mustResource(t, s, "Sheet", domain.KindMaterial, "kg", floatPtr(1))
	mustProcess(t, s, "Cut", nil, ProcessResourceInput{ResourceID: sheet.ID, Quantity: 1})
	fits := mustOrder(t, s, "CAP-1", 1)
	over := mustOrder(t, s, "CAP-2", 1)

	if _, err := s.StartStep(ctx, fits.ID, stepID(t, fits, 1)); err != nil {
		t.Fatalf("start order that fits: %v", err)
	}
	if got := consumed(t, s, sheet.ID); got != 1 {
		t.Fatalf("consumed after start = %v, want 1", got)
	}

	_, err := s.StartStep(ctx, over.ID, stepID(t, over, 1))
	wantsCode(t, err, "insufficient_resource")
	if got := consumed(t, s, sheet.ID); got != 1 {
		t.Fatalf("rejected start must not consume: got %v, want 1", got)
	}
}

// Consumption is recorded when the work starts, is not double counted when the
// step completes, and the order is done once every step is done.
func TestConsumptionRecordedOnStartAndOrderCompletes(t *testing.T) {
	s := newService(t)
	sheet := mustResource(t, s, "Sheet", domain.KindMaterial, "kg", nil)
	mustProcess(t, s, "Cut", nil, ProcessResourceInput{ResourceID: sheet.ID, Quantity: 2})
	order := mustOrder(t, s, "LED-1", 3)
	step := stepID(t, order, 1)

	if _, err := s.StartStep(ctx, order.ID, step); err != nil {
		t.Fatalf("start: %v", err)
	}
	if got := consumed(t, s, sheet.ID); got != 6 {
		t.Fatalf("consumed while running = %v, want 6", got)
	}

	updated, err := s.CompleteStep(ctx, order.ID, step)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if got := consumed(t, s, sheet.ID); got != 6 {
		t.Fatalf("consumed after complete = %v, want 6 (no double count)", got)
	}
	if updated.Status != domain.OrderDone {
		t.Fatalf("order status = %q, want %q", updated.Status, domain.OrderDone)
	}
}

// Cancelling is only allowed before any work has started: once a step has run,
// its consumption is committed and the order has to be finished instead.
func TestCancelRefusedOnceStarted(t *testing.T) {
	s := newService(t)
	sheet := mustResource(t, s, "Sheet", domain.KindMaterial, "kg", nil)
	mustProcess(t, s, "Cut", nil, ProcessResourceInput{ResourceID: sheet.ID, Quantity: 1})
	mustProcess(t, s, "Weld", nil, ProcessResourceInput{ResourceID: sheet.ID, Quantity: 1})
	started := mustOrder(t, s, "CAN-1", 1)
	parked := mustOrder(t, s, "CAN-2", 1)
	pending := mustOrder(t, s, "CAN-3", 1)

	if _, err := s.StartStep(ctx, started.ID, stepID(t, started, 1)); err != nil {
		t.Fatalf("start: %v", err)
	}
	_, err := s.CancelOrder(ctx, started.ID)
	wantsCode(t, err, "order_started")

	// Finished its Cut and waiting at Weld: started, so still refused.
	if _, err := s.StartStep(ctx, parked.ID, stepID(t, parked, 1)); err != nil {
		t.Fatalf("start parked: %v", err)
	}
	if _, err := s.CompleteStep(ctx, parked.ID, stepID(t, parked, 1)); err != nil {
		t.Fatalf("complete parked: %v", err)
	}
	_, err = s.CancelOrder(ctx, parked.ID)
	wantsCode(t, err, "order_started")

	cancelled, err := s.CancelOrder(ctx, pending.ID)
	if err != nil {
		t.Fatalf("cancel pending order: %v", err)
	}
	if cancelled.Status != domain.OrderCancelled {
		t.Fatalf("status = %q, want %q", cancelled.Status, domain.OrderCancelled)
	}
}

// A resource that a process recipe depends on cannot be deleted.
func TestResourceInUseCannotBeDeleted(t *testing.T) {
	s := newService(t)
	sheet := mustResource(t, s, "Sheet", domain.KindMaterial, "kg", nil)
	mustProcess(t, s, "Cut", nil, ProcessResourceInput{ResourceID: sheet.ID, Quantity: 1})

	err := s.DeleteResource(ctx, sheet.ID)
	wantsCode(t, err, "resource_in_use")
}

// The WIP and capacity checks are only sound because starting a step takes row
// locks on the process it runs and on the resources it consumes. Prove those
// locks are really held: while another transaction owns the row, a start waits.
func TestStartStepWaitsForConflictingLocks(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, s *Service) *domain.Order
		hold  func(t *testing.T, tx *gorm.DB)
		// want is the error code a start must end with once the lock is
		// released; "" means it must succeed.
		want string
	}{
		{
			name: "process row",
			setup: func(t *testing.T, s *Service) *domain.Order {
				mustProcess(t, s, "Gate", intPtr(1))
				return mustOrder(t, s, "LOCK-P", 1)
			},
			hold: func(t *testing.T, tx *gorm.DB) {
				var id int64
				if err := tx.Raw("SELECT id FROM processes WHERE name = ? FOR UPDATE", "Gate").Scan(&id).Error; err != nil {
					t.Fatalf("lock the process row: %v", err)
				}
				if id == 0 {
					t.Fatal("no process row was locked")
				}
			},
		},
		{
			name: "resource row",
			setup: func(t *testing.T, s *Service) *domain.Order {
				sheet := mustResource(t, s, "Sheet", domain.KindMaterial, "kg", floatPtr(10))
				mustProcess(t, s, "Cut", nil, ProcessResourceInput{ResourceID: sheet.ID, Quantity: 1})
				return mustOrder(t, s, "LOCK-R", 1)
			},
			// Shrink the resource while holding it. A start that takes the row
			// lock before checking capacity sees 0.5 and must refuse; one that
			// reads it unlocked sails past the check and overspends the stock.
			hold: func(t *testing.T, tx *gorm.DB) {
				var id int64
				if err := tx.Raw("UPDATE resources SET capacity = 0.5 WHERE name = ? RETURNING id", "Sheet").Scan(&id).Error; err != nil {
					t.Fatalf("shrink the resource: %v", err)
				}
				if id == 0 {
					t.Fatal("no resource row was updated")
				}
			},
			want: "insufficient_resource",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newService(t)
			order := tc.setup(t, s)
			step := stepID(t, order, 1)

			// Another transaction takes the row first.
			holder := s.db.Begin()
			if holder.Error != nil {
				t.Fatalf("begin holder transaction: %v", holder.Error)
			}
			defer func() { _ = holder.Rollback().Error }()
			tc.hold(t, holder)

			done := make(chan error, 1)
			go func() {
				_, err := s.StartStep(ctx, order.ID, step)
				done <- err
			}()

			select {
			case err := <-done:
				t.Fatalf("start finished while another transaction held the %s (err=%v)", tc.name, err)
			case <-time.After(300 * time.Millisecond):
			}

			// Releasing it must let the start through.
			if err := holder.Commit().Error; err != nil {
				t.Fatalf("commit holder transaction: %v", err)
			}
			select {
			case err := <-done:
				if tc.want == "" {
					if err != nil {
						t.Fatalf("start after the lock was released: %v", err)
					}
					return
				}
				wantsCode(t, err, tc.want)
			case <-time.After(5 * time.Second):
				t.Fatal("start was still blocked after the lock was released")
			}
		})
	}
}

// boardColumn finds the board column for a process by name.
func boardColumn(t *testing.T, board *Board, name string) BoardColumn {
	t.Helper()
	for _, column := range board.Columns {
		if column.Process.Name == name {
			return column
		}
	}
	t.Fatalf("board has no column for process %q", name)
	return BoardColumn{}
}

func references(steps []BoardStep) []string {
	out := make([]string, 0, len(steps))
	for _, step := range steps {
		out = append(out, step.Reference)
	}
	return out
}

// The board shows what runs at each step and what is next in line for it: an
// order appears only in the column of its current step, never further down its
// route, and a queued step says whether it can start right now.
func TestBoardListsQueuedAndFlagsStartable(t *testing.T) {
	s := newService(t)
	sheet := mustResource(t, s, "Sheet", domain.KindMaterial, "kg", nil)
	mustProcess(t, s, "Cut", nil, ProcessResourceInput{ResourceID: sheet.ID, Quantity: 1})
	mustProcess(t, s, "Weld", intPtr(1), ProcessResourceInput{ResourceID: sheet.ID, Quantity: 1})

	first := mustOrder(t, s, "B-1", 1)
	second := mustOrder(t, s, "B-2", 1)
	mustOrder(t, s, "B-3", 1)

	board, err := s.Board(ctx)
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	cut := boardColumn(t, board, "Cut")
	weld := boardColumn(t, board, "Weld")

	if got, want := references(cut.Queued), []string{"B-1", "B-2", "B-3"}; !slices.Equal(got, want) {
		t.Errorf("Cut queue = %v, want %v (oldest order first)", got, want)
	}
	if !cut.Queued[0].Startable {
		t.Error("the first queued Cut step should be startable")
	}
	// Every order is still at Cut, so none of them belongs in Weld yet.
	if len(weld.Queued) != 0 {
		t.Errorf("Weld queue = %v, want empty while every order is still at Cut", references(weld.Queued))
	}

	// Finishing the first order's Cut moves it — and only it — into Weld.
	if _, err := s.StartStep(ctx, first.ID, stepID(t, first, 1)); err != nil {
		t.Fatalf("start Cut: %v", err)
	}
	if _, err := s.CompleteStep(ctx, first.ID, stepID(t, first, 1)); err != nil {
		t.Fatalf("complete Cut: %v", err)
	}

	board, err = s.Board(ctx)
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	if got, want := references(boardColumn(t, board, "Cut").Queued), []string{"B-2", "B-3"}; !slices.Equal(got, want) {
		t.Errorf("Cut queue = %v, want %v once B-1 moved on", got, want)
	}
	weld = boardColumn(t, board, "Weld")
	if got, want := references(weld.Queued), []string{"B-1"}; !slices.Equal(got, want) {
		t.Errorf("Weld queue = %v, want %v (only the order that reached Weld)", got, want)
	}
	if !weld.Queued[0].Startable {
		t.Error("B-1 should be startable at Weld once its Cut is done")
	}

	// Weld runs one order at a time. Take the slot, then let the next order
	// reach Weld: it queues up but cannot start.
	if _, err := s.StartStep(ctx, first.ID, stepID(t, first, 2)); err != nil {
		t.Fatalf("start Weld: %v", err)
	}
	if _, err := s.StartStep(ctx, second.ID, stepID(t, second, 1)); err != nil {
		t.Fatalf("start B-2 Cut: %v", err)
	}
	if _, err := s.CompleteStep(ctx, second.ID, stepID(t, second, 1)); err != nil {
		t.Fatalf("complete B-2 Cut: %v", err)
	}

	board, err = s.Board(ctx)
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	weld = boardColumn(t, board, "Weld")
	if got, want := references(weld.Running), []string{"B-1"}; !slices.Equal(got, want) {
		t.Errorf("Weld running = %v, want %v", got, want)
	}
	if got, want := references(weld.Queued), []string{"B-2"}; !slices.Equal(got, want) {
		t.Errorf("Weld queue = %v, want %v", got, want)
	}
	if weld.Queued[0].Startable {
		t.Error("B-2 should wait while B-1 holds the only Weld slot")
	}
}
