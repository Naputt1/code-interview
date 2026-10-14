// Package domain holds the core entities of the manufacturing tracker.
//
// The process sequence is data, not code: resources and processes are created
// by the user at runtime, so nothing about "Step A -> Step B -> Step C" is
// hardcoded. An order snapshots the current sequence into order_steps when it
// is created, so later edits to the process list never rewrite history.
package domain

import "time"

// Resource kinds. A resource is anything a step consumes: stock material,
// machine time, or labour.
const (
	KindMaterial = "material"
	KindMachine  = "machine"
	KindLabour   = "labour"
)

// Order status values.
const (
	OrderPending    = "pending"
	OrderInProgress = "in_progress"
	OrderDone       = "done"
	OrderCancelled  = "cancelled"
)

// Order step status values.
const (
	StepQueued  = "queued"
	StepRunning = "running"
	StepDone    = "done"
)

// Resource is a consumable used by one or more processes. Capacity is the
// total stock available for the run; a nil capacity means the resource is
// effectively unlimited (typical for labour).
type Resource struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null;uniqueIndex" json:"name"`
	Kind      string    `gorm:"not null" json:"kind"`
	Unit      string    `gorm:"not null" json:"unit"`
	Capacity  *float64  `gorm:"column:capacity" json:"capacity"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Resource) TableName() string { return "resources" }

// Process is one working step in the sequence. Position defines its order in
// the sequence; wipLimit caps how many orders may run at this step at once
// (nil means unlimited).
type Process struct {
	ID        int64             `gorm:"primaryKey" json:"id"`
	Name      string            `gorm:"not null;uniqueIndex" json:"name"`
	Position  int               `gorm:"not null;index" json:"position"`
	WIPLimit  *int              `gorm:"column:wip_limit" json:"wipLimit"`
	Inputs    []ProcessResource `gorm:"foreignKey:ProcessID" json:"inputs"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

func (Process) TableName() string { return "processes" }

// ProcessResource is one line of a process's recipe: how much of a resource a
// single unit of work at this step consumes.
type ProcessResource struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	ProcessID  int64     `gorm:"not null;index;uniqueIndex:idx_process_resource" json:"processId"`
	ResourceID int64     `gorm:"not null;index;uniqueIndex:idx_process_resource" json:"resourceId"`
	Quantity   float64   `gorm:"not null" json:"quantity"`
	Resource   *Resource `gorm:"foreignKey:ResourceID" json:"resource,omitempty"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

func (ProcessResource) TableName() string { return "process_resources" }

// Order is a job moving through the process sequence.
type Order struct {
	ID        int64       `gorm:"primaryKey" json:"id"`
	Reference string      `gorm:"not null" json:"reference"`
	Quantity  int         `gorm:"not null" json:"quantity"`
	Status    string      `gorm:"not null;index" json:"status"`
	Steps     []OrderStep `gorm:"foreignKey:OrderID" json:"steps,omitempty"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

func (Order) TableName() string { return "orders" }

// OrderStep is one order's run of one process. Several order_steps may be
// running against the same process at once: that is what "a step handles one
// or more orders concurrently" means in the data model.
type OrderStep struct {
	ID           int64         `gorm:"primaryKey" json:"id"`
	OrderID      int64         `gorm:"not null;index" json:"orderId"`
	ProcessID    int64         `gorm:"not null;index" json:"processId"`
	Position     int           `gorm:"not null" json:"position"`
	Status       string        `gorm:"not null;index" json:"status"`
	StartedAt    *time.Time    `json:"startedAt"`
	FinishedAt   *time.Time    `json:"finishedAt"`
	Process      *Process      `gorm:"foreignKey:ProcessID" json:"process,omitempty"`
	Consumptions []Consumption `gorm:"foreignKey:OrderStepID" json:"consumptions,omitempty"`
	CreatedAt    time.Time     `json:"-"`
	UpdatedAt    time.Time     `json:"-"`
}

func (OrderStep) TableName() string { return "order_steps" }

// Consumption is an append-only ledger entry recording that an order step drew
// a quantity of a resource. Everything the UI shows about resource use is a
// SUM over this table, so consumption can never drift from reality.
type Consumption struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	OrderStepID int64     `gorm:"not null;index" json:"orderStepId"`
	ResourceID  int64     `gorm:"not null;index" json:"resourceId"`
	Quantity    float64   `gorm:"not null" json:"quantity"`
	OccurredAt  time.Time `gorm:"not null" json:"occurredAt"`
}

func (Consumption) TableName() string { return "consumptions" }
