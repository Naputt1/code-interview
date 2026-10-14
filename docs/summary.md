# Written summary

## 1. Summary of the requirements

The prototype makes a manufacturing process observable. Work runs through a defined
sequence of steps, each of which may handle several orders at once and consume
resources such as materials, machine time or labour. A user must create orders and
advance them through the steps, see what is in progress at each, watch consumption as
it happens, and overview every order. The assignment leaves the definitions, rules and
UI open, and weights process reasoning above features.

## 2. Turning the requirement into a functional specification

Steps and resources are data: the user creates both, and nothing in the code assumes a
particular line, so **Setup** defines one at runtime. A **resource** has a name, kind,
unit and an optional capacity, where empty means unlimited. A **process** is one step
with a position, an optional WIP limit and a per-unit recipe. An **order** holds a
reference and a quantity; creating one copies the line into its steps so later edits do
not rewrite work in flight. An **order step** sends one order through one process via
queued, running and done, and several running at one process is the concurrency the
assignment asks for. A **consumption** row is the append-only ledger, and every UI
figure aggregates over it.

The rules:

1. **Sequence.** A step starts only once every earlier step is done.
2. **Consumption on start.** Starting a step writes recipe × quantity, so consumption
   is visible during the work.
3. **Capacity.** A start exceeding a finite resource's remaining stock is refused
   before anything is written, and capacity cannot be set below what has already been
   consumed.
4. **WIP limit.** A start is refused while the process already runs its limit.
5. **One running step per order**, done once every step is done.
6. **Cancellation** only before any step has started.
7. **Referential integrity.** A resource or process still referenced by a recipe or a
   ledger row cannot be deleted.

The four scope bullets map onto three screens: **Setup** defines the line, **Orders**
creates and advances, and **Board** shows progress, the queue behind each step, the
consumption and the overview.

## 3. Software design

A modular monolith. All rules live in `service`, the only code that writes. Starting a
step runs in one transaction that locks the order, the process and its resources in a
fixed order, so two starts racing for the last WIP slot cannot both get in. `httpapi`
is a thin Gin layer mapping domain errors to status codes.

Swagger annotations generate a committed OpenAPI spec, and Orval generates the typed
client and Query hooks from it. Every call passes through one `request` helper, and one
`src/api` module owns the query keys, poll interval and invalidation. TanStack Router
navigates, TanStack Query holds server state, Zustand holds view preferences. The rules
are tested against a real PostgreSQL, since locking, uniqueness and capacity arithmetic
are database behaviour.

Left out on purpose: authentication, since the assignment is about the process rather
than who runs it; streaming, since the board polls; and auto-migration, since numbered
SQL files keep the schema history readable.

## 4. Simplified user manual

1. **Setup, Resources.** Give each resource a kind, a unit, and a capacity only if it
   is finite.
2. **Setup, Processes.** Add the steps in order, each with a WIP limit and a recipe,
   for example `Cut` consuming `Steel ×0.5` per unit.
3. **Orders, New order.** Enter a reference and a quantity. The app creates one queued
   step per process.
4. **Work the order.** Only the next step can start. **Start** records consumption
   immediately and **Complete** releases the next step. Cancel is offered wherever an
   order is listed, and only before it starts.
5. **Board.** Counters cover the whole order book, and each column shows what is being
   worked there with the queue behind it. A waiting order starts from the board unless
   the WIP limit is full, and the resource panel shows consumed against capacity.

A refused action says why, for example "Paint can only run 1 order(s) at a time".

## 5. Summary of the software

Go and PostgreSQL behind a Gin API, with the contract generated from the handlers and
the React client generated from the contract. The line is user-defined, orders move
through a strictly enforced sequence, and every unit of consumption is a ledger row the
UI aggregates into progress and load.

Known limits: no authentication or per-user scoping; the board polls rather than
streams; consumption is committed at step start, so a started step that is never
finished stays on the books; and steps cannot be skipped or partially completed.
