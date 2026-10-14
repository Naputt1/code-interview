# Written summary

## 1. Summary of the requirements

The prototype must make a manufacturing process observable. The process runs
through a defined sequence of working steps; each step may handle several orders
concurrently; each step consumes resources such as materials, machine time or
labour. A user must be able to create orders and advance them through the steps,
see which orders are in progress at each step, watch resource consumption as work
progresses, and get an overview of every order. The assignment leaves the
definitions of steps, orders, resources, their interaction rules and the UI open,
and states that reasoning about the process as a whole matters more than any
single feature.

## 2. Turning the requirement into a functional specification

The two definitions the assignment leaves open — "steps" and "resources" — are
treated as **configuration, not code**. Rather than inventing one line and
hard-coding it, the platform ships a **Setup** screen: a user defines their own
line at runtime. That single decision is what makes the rest of the specification
expressible.

The model:

- **Resource** — a name, a kind (material / machine / labour) and a unit.
  Capacity is optional; leaving it empty means effectively unlimited (labour, or
  a machine that is not tracked by the hour). Setting it makes the resource
  exhaustible and therefore worth guarding.
- **Process** (a step) — a named workstation at a `position` in the sequence, an
  optional WIP limit (how many orders it may work at once) and a **recipe**: a set
  of resources, each with a quantity per unit of product.
- **Order** — a reference and a quantity. Creating an order *snapshots* the
  current line into that order's steps, so later edits to the line do not
  retroactively rewrite work already in flight.
- **Order step** — one order at one process, `queued → running → done`. Several
  running steps on the same process is precisely the assignment's "a step may
  handle one or more orders concurrently".
- **Consumption** — an append-only ledger row recording which order step consumed
  which resource, how much, and when. Every consumption number in the UI is an
  aggregate over that ledger.

The rules, which are the actual process reasoning:

1. **Sequence** — a step can start only when every earlier step on that order is
   done.
2. **Consumption on start** — starting a step records recipe × order quantity, so
   consumption is visible *while* work is in progress, not just afterwards.
3. **Capacity** — a start that would exceed a finite resource's remaining
   capacity is refused before anything is written, so the ledger never records a
   partially applied step.
4. **WIP limit** — a start is refused when the process already holds its limit of
   running orders.
5. **One running step per order** — enforced at start, and an order becomes `done`
   when every step is done.
6. **Cancellation** — an order can be cancelled only before any of its steps has
   started, so cancelling never strands already-committed resources; an order
   that has begun work must be carried through. A resource or process that is
   still referenced by a recipe or a ledger entry cannot be deleted.

The four scope bullets then map onto three screens: **Setup** (define the line),
**Orders** (create and advance), **Board** (in-progress per step plus the queue
behind it, consumption, and the status overview).

## 3. Software design

A modular monolith with a contract-first seam.

- **Go API** — `domain` holds the GORM models and a typed `Error` (status, code,
  message). `service` owns every rule and is the only writer to the database;
  order transitions run in one transaction that locks the order, the target
  process and the resources it consumes (always in the same order), so two
  concurrent starts cannot both pass the WIP or capacity checks. `httpapi` is a
  thin Gin layer that binds, delegates, and maps domain errors to HTTP codes;
  `store` owns the connection and a small embedded migration runner.
- **Contract** — swaggo annotations on the handlers generate the OpenAPI spec,
  which is committed. This keeps the document adjacent to the code that serves it
  and gives the frontend a single machine-readable source.
- **Frontend** — Orval turns that spec into the typed client and TanStack Query
  hooks; all calls pass through one hand-written `request` helper so error
  handling exists once, and the generated code is wrapped in a single `src/api`
  module that owns the query keys, the poll interval and the invalidation, so a
  step change refreshes exactly the board, orders and ledger it affects. TanStack
  Router owns navigation; TanStack Query owns server state (cached, invalidated,
  polled); Zustand holds only client-side view preferences. The board polls every
  four seconds, with a pause control.
- **Testing** — the rules are tested at the service level against a real
  PostgreSQL, because the guarantees that matter (locking, uniqueness, the
  capacity arithmetic) are database behaviour.

Deliberate exclusions, and why: **no authentication** (the assignment is about
the process, not identity); **no websocket/SSE** (polling keeps the demo
dependency-free while the "live" behaviour is still observable); **no ORM
auto-migration in production paths** (numbered SQL migrations are the record).

## 4. Simplified user manual

1. **Setup → Resources.** Create what the line consumes. Give it a kind, a unit,
   and a capacity only if it is finite (e.g. `Steel, material, kg, 100`; leave
   capacity empty for labour).
2. **Setup → Processes.** Add the steps in order and give each one a WIP limit and
   a recipe — the resources it consumes per unit (e.g. `Cut` consumes
   `Steel ×0.5`). Steps run top to bottom in the order listed.
3. **Orders → New order.** Enter a reference and a quantity. The order is created
   with one queued step per process.
4. **Open the order and work it.** Only the next step in the sequence can be
   started. **Start** records consumption immediately; **Complete** finishes the
   step and releases the next one. The order is `done` when all steps are done.
   **Cancel order** is available while no step is running.
5. **Board.** Read the state of the plant: counters for the whole order book, one
   column per process showing the orders being worked there (with elapsed time
   and a one-click Complete) and, beneath them, the orders waiting at that step
   in line order. An order appears only at the step it has actually reached, so a
   column shows what is in progress there and what is next for it — never the
   steps further down an order's route. A waiting order can be started straight
   from the board unless the step's WIP limit is full, in which case the card says
   so. The resource panel shows consumed against capacity, and auto-refresh can
   be paused.

Starting a step that would break a rule is refused with a specific message — for
example "Paint can only run 1 order(s) at a time" or "not enough Steel remaining"
— so the operator learns the constraint rather than seeing a silent failure.

## 5. Summary of the software

A working prototype: Go + PostgreSQL + Gin behind a generated OpenAPI contract,
and a React + TanStack front end whose client is generated from that contract.
The line is user-defined, orders advance through a strictly enforced sequence,
and every unit of resource consumption is an auditable ledger row that the UI
aggregates into progress and load. The rules are unit-tested against a real
database, and the whole stack is exercised end to end in the browser.

Known limits, stated plainly: there is no authentication or per-user scoping; the
board polls rather than streaming; consumption is recorded when a step starts
(committed) rather than when it finishes, so a step that is started and never
completed stays on the books; and steps cannot be skipped, reordered or partially
completed.
