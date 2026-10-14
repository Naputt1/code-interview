# Written summary

## 1. Summary of the requirements

The prototype must make a manufacturing process observable: work runs through a
defined sequence of steps, each step may handle several orders at once, and each
consumes resources such as materials, machine time or labour. A user must create
orders and advance them through the steps, see what is in progress at each step,
watch consumption as work progresses, and get an overview of every order. The
assignment leaves the definitions, the rules and the UI open, and weights
reasoning about the process above any single feature.

## 2. Turning the requirement into a functional specification

"Steps" and "resources" are **configuration, not code**: a **Setup** screen
defines the line at runtime, and nothing about the sequence is hard-coded.

- **Resource** — name, kind (material / machine / labour), unit; capacity is
  optional, empty meaning unlimited.
- **Process** (a step) — a workstation at a `position`, an optional WIP limit, and
  a **recipe** of resources with a quantity per unit.
- **Order** — reference and quantity; creating one *snapshots* the line into its
  steps, so later edits do not rewrite work already in flight.
- **Order step** — one order at one process, `queued → running → done`; several
  running steps on one process is the assignment's "one or more orders
  concurrently".
- **Consumption** — an append-only ledger row of which step used which resource,
  how much and when; every UI figure aggregates over it.

The rules:

1. **Sequence** — a step starts only once every earlier step on that order is done.
2. **Consumption on start** — starting records recipe × quantity, so consumption
   is visible during the work, not only after it.
3. **Capacity** — a start exceeding a finite resource's remaining capacity is
   refused before anything is written, so no partial step reaches the ledger.
4. **WIP limit** — a start is refused while the process already runs its limit.
5. **One running step per order** — an order is `done` when every step is.
6. **Cancellation** — only before any step has started, so cancelling never strands
   committed resources; anything still referenced by a recipe or ledger entry
   cannot be deleted.

The four scope bullets map onto three screens: **Setup** (define the line),
**Orders** (create and advance), **Board** (in-progress per step, the queue behind
it, consumption and the overview).

## 3. Software design

A modular monolith with a contract-first seam.

- **Go API** — `service` owns every rule and is the only writer; order transitions
  run in one transaction that locks the order, the process and its resources in a
  fixed order, so two concurrent starts cannot both pass the WIP or capacity
  checks. `httpapi` is a thin Gin layer mapping domain errors to HTTP codes.
- **Contract** — swaggo annotations generate the committed OpenAPI spec, giving the
  frontend one machine-readable source.
- **Frontend** — Orval generates the typed client and Query hooks; all calls pass
  through one `request` helper, and a single `src/api` module owns the keys, the
  poll interval and the invalidation. Router owns navigation, Query owns server
  state, Zustand only view preferences.
- **Testing** — the rules are tested against a real PostgreSQL, because the
  guarantees that matter (locking, uniqueness, capacity arithmetic) are database
  behaviour.

Excluded on purpose: **authentication** (the assignment is about the process, not
identity), **streaming** (polling keeps the demo dependency-free while "live"
behaviour stays observable), **ORM auto-migration** (numbered SQL migrations are
the record).

## 4. Simplified user manual

1. **Setup → Resources.** Kind, unit, and a capacity only if it is finite (empty
   for labour).
2. **Setup → Processes.** Steps in order, each with a WIP limit and a recipe (e.g.
   `Cut` consumes `Steel ×0.5`).
3. **Orders → New order.** A reference and a quantity; one queued step per process.
4. **Work the order.** Only the next step can start. **Start** records consumption
   immediately, **Complete** releases the next one. Cancel is offered wherever an
   order is listed, and only before it starts.
5. **Board.** Counters for the order book; per process, what is being worked there
   and the queue behind it. An order appears only at the step it has reached, and a
   waiting order starts straight from the board unless the WIP limit is full. The
   resource panel shows consumed against capacity.

Breaking a rule is refused with a specific message — "Paint can only run 1 order(s)
at a time" — so the constraint is learned rather than silently failing.

## 5. Summary of the software

Go + PostgreSQL + Gin behind a generated OpenAPI contract, with a React + TanStack
front end whose client is generated from it. The line is user-defined, orders
advance through a strictly enforced sequence, and every unit of consumption is an
auditable ledger row the UI aggregates into progress and load. Known limits: no
authentication or per-user scoping; polling rather than streaming; consumption is
committed at step start, so a started-but-unfinished step stays on the books; and
steps cannot be skipped or partially completed.
