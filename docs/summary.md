# Written summary

## 1. Summary of the requirements

The prototype has to make a manufacturing process observable. Work runs through a
defined sequence of steps, each of which may handle several orders at once and
consume resources such as materials, machine time or labour. A user must
create orders and advance them through the steps, see what is in progress at each
step, watch consumption as work progresses, and get an overview of every order.
The assignment leaves the definitions, the rules and the UI open, and weights
reasoning about the process above any single feature.

## 2. Turning the requirement into a functional specification

Steps and resources are records the user creates in the app. Nothing in the code
assumes a particular line, so **Setup** can define one at runtime.

- **Resource.** A name, a kind (material, machine or labour) and a unit. Capacity
  is optional, and an empty capacity means unlimited.
- **Process.** A single step, with a position in the sequence, an optional WIP
  limit, and a recipe of the resources it consumes per unit.
- **Order.** A reference and a quantity. Creating one copies the current line into
  its steps, so later edits do not rewrite work already in flight.
- **Order step.** One order at one process, moving through queued, running and
  done. Several steps running at the same process is the assignment's "one or more
  orders concurrently".
- **Consumption.** An append-only ledger of which step used which resource, how
  much and when. Every UI figure aggregates over it.

The rules:

1. **Sequence.** A step starts only once every earlier step on that order is done.
2. **Consumption on start.** Starting a step writes recipe × order quantity, so
   consumption is visible during the work, not only after it.
3. **Capacity.** A start that would exceed a finite resource's remaining stock is
   refused before anything is written, so no half-recorded step reaches the ledger.
4. **WIP limit.** A start is refused while the process is already running its
   limit.
5. **One running step per order.** An order becomes done when every step is done.
6. **Cancellation.** An order can be cancelled only before any step has started, so
   cancelling never strands committed resources.
7. **Referential integrity.** A resource or process still referenced by a recipe or
   a ledger row cannot be deleted.

The four scope bullets map onto three screens. **Setup** defines the line,
**Orders** creates and advances, and **Board** shows what is in progress at each
step with the queue behind it, the consumption and the overview.

## 3. Software design

The backend is a modular monolith.

- **Service layer.** All the rules live in `service`, which is the only code that
  writes. Starting a step runs in one transaction that locks the order, the process
  and the resources it consumes, always in the same order, so two starts racing for
  the last WIP slot cannot both get in. `httpapi` is a thin Gin layer that maps
  domain errors to HTTP status codes.
- **Contract.** Swagger annotations in the handlers generate the OpenAPI spec,
  which is committed, so the frontend has one machine-readable source instead of
  hand-written request types.
- **Frontend.** Orval generates the typed client and the Query hooks from that
  spec. Every call goes through one `request` helper, and a single `src/api` module
  owns the query keys, the poll interval and the invalidation. TanStack Router
  handles navigation, TanStack Query holds server state, and Zustand holds only
  view preferences.
- **Testing.** The rules are tested against a real PostgreSQL. Locking, uniqueness
  and the capacity arithmetic are database behaviour, so an in-memory double would
  not prove anything.

Three things are deliberately absent. There is no authentication, since the
assignment is about the process rather than who runs it. The board polls instead of
streaming, which keeps the demo to one dependency and still shows the state
changing. Migrations are numbered SQL files, so the schema history is readable.

## 4. Simplified user manual

1. **Setup, Resources.** Give each resource a kind and a unit. Add a capacity only
   if it is finite, and leave it empty for things like labour.
2. **Setup, Processes.** Add the steps in order. Each one gets a WIP limit and a
   recipe, for example `Cut` consuming `Steel ×0.5` per unit.
3. **Orders, New order.** Enter a reference and a quantity. The app creates one
   queued step per process.
4. **Work the order.** Only the next step can start. **Start** records the
   consumption straight away and **Complete** releases the following step. Cancel
   is offered wherever an order is listed, and only while it has not started.
5. **Board.** Counters show the whole order book. Each process column shows what is
   being worked there and the queue behind it. An order appears only at the step it
   has reached, and a waiting order can be started straight from the board unless
   the WIP limit is full. The resource panel shows consumed against capacity.

A refused action says why, for example "Paint can only run 1 order(s) at a time".

## 5. Summary of the software

Go and PostgreSQL behind a Gin API. The OpenAPI contract is generated from the
handlers and the React client is generated from the contract. The
line is defined by the user, orders move through a strictly enforced sequence, and
every unit of consumption is a ledger row that the UI aggregates into progress and
load.

Known limits: there is no authentication or per-user scoping; the board polls
rather than streaming; consumption is committed when a step starts, so a step that
was started and never finished stays on the books; and steps cannot be skipped or
partially completed.
