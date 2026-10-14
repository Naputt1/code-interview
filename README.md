# Manufacturing Process Tracker

A prototype that tracks orders through a multi-step manufacturing process and shows
the resources each step consumes. Built for the Meticuly preliminary assignment.

The written submission is in [`docs/summary.md`](docs/summary.md); how AI tools
were used is in [`docs/ai-usage.md`](docs/ai-usage.md).

## What it does

| Scope (from the assignment) | Where in the app |
| --- | --- |
| create orders and advance them through the defined steps | **Orders** to create one, **order detail** to Start and Complete each step |
| view the orders currently in progress at each step | **Board**, one column per process showing the orders running there and the queue behind them |
| observe resource consumption as work progresses | **Board**, the Resource consumption panel updated as steps start |
| obtain an overview of the current status of all orders | **Board**, status counters and per-process load; **Orders**, the full table |

Which steps exist, in what order, and what each one consumes is defined by the user
in **Setup**. The assignment leaves that open, and nothing about the sequence is
hard-coded.

## Stack

- **Backend.** Go 1.25, Gin, GORM and PostgreSQL. Swagger annotations in the
  handlers generate the OpenAPI contract.
- **Frontend.** React 19 on Vite, TanStack Router, TanStack Query, Orval for the
  generated client and hooks, shadcn/ui on Tailwind v4, and Zustand for client-only
  UI state.
- **Dev.** [fog](https://github.com/Naputt1/fog) runs the API and the web app
  locally. Docker Compose, or a Postgres you already have, provides the database.

## How it works

Steps and resources are rows in the database; the code never assumes a particular
line. The sequence is the `position` column on `processes`, each step's recipe is
the `process_resources` join, and every unit of consumption is a row in an
append-only `consumptions` ledger keyed to the `order_step` that caused it. Every
consumption figure in the UI is an aggregate over that ledger, so the board and the
order view cannot disagree.

The rules for sequencing, WIP limits, capacity, one running step per order and
cancellation all live in one transactional service layer
(`backend/internal/service`), which is what the tests exercise.

See [`docs/summary.md`](docs/summary.md) for the reasoning behind these choices.

## Layout

```
backend/
  cmd/api/main.go            # entry point, migrations + optional seed on start
  internal/domain/           # GORM models, domain errors
  internal/service/          # the rules (orders.go, views.go, service.go) + tests
  internal/httpapi/          # Gin handlers, DTOs, error mapping
  internal/store/            # connection + embedded migration runner
  migrations/                # 000001_init.{up,down}.sql, embedded
  api/swagger/openapi3.yaml  # committed contract, generated from annotations
  Makefile                   # run / build / test / vet / fmt / check / swagger
frontend/
  orval.config.ts            # contract -> typed TanStack Query client
  src/api/                   # client.ts + mutator.ts (hand-written transport),
                             # gen/ (generated) and the keys/queries/mutations
                             # module the app imports instead of gen/ directly
  src/routes/                # TanStack Router file routes (board, orders, setup)
  src/features/              # board/, orders/, processes/, resources/
  src/components/ui/         # shadcn primitives
  src/state/                 # zustand store (client-only prefs)
infra/docker-compose.yml     # Postgres for local development
fog.json                     # local dev orchestration
```

## Getting started

You need Go 1.25+, Node 24+ with pnpm, and PostgreSQL 16.

### Option A: fog (how the project is developed)

```sh
fog dev
```

`fog.json` starts the API with hot reload, then the Vite dev server wired to it, and
exposes them at `https://<branch>.api.code-interview` and
`https://<branch>.code-interview`. It expects a reachable Postgres, and the
`DATABASE_URL` in `fog.json` points at a local instance. To use the Docker database
instead, run `fog infra` and point `DATABASE_URL` at it. Ports are assigned
dynamically, so read them from `fog ls` or `fog logs <pid>`.

Run fog from inside a checkout. It names each instance after its git branch
(`${branch}` in the endpoint hosts) and aborts with `requires a git branch (not in
a worktree)` when it cannot resolve one. Setting up DNS and the router needs sudo
the first time.

### Option B: Docker Postgres, two terminals

```sh
docker compose -f infra/docker-compose.yml up -d   # Postgres 16 on :5432

cd backend && make run                             # http://localhost:8080
cd frontend && pnpm install && pnpm dev            # http://localhost:5173
```

The database defaults (`tracker/tracker/tracker`) match the Compose service, and
the Vite dev server proxies `/api` to the API. Both default to port 8080. If fog's
router is already running it holds that port for its dashboard, so start the API
elsewhere and point the dev server at it:

```sh
SERVER_ADDR=:8090 make run                        # in backend
API_PROXY_TARGET=http://127.0.0.1:8090 pnpm dev   # in frontend
```

### Option C: an existing Postgres, no Docker

```sh
createdb code_interview
cd backend
DATABASE_URL='postgres://<user>@127.0.0.1:5432/code_interview?sslmode=disable' make run
cd frontend && pnpm install && pnpm dev
```

Migrations run automatically at API startup. The first time the database has no
processes, the app seeds a demo line of Cut, Weld, Paint and Assemble with orders
spread across it. Set `SEED_IF_EMPTY=false` to start empty.

## The API contract

The swaggo annotations in the handlers are the source of truth. Regenerating is
deterministic, and both artefacts are committed so a checkout builds without
running the generators:

```sh
cd backend  && make swagger     # -> backend/api/swagger/openapi3.yaml
cd frontend && pnpm generate    # orval -> src/api/gen (client + Query hooks)
```

`customInstance` (`src/api/mutator.ts`) routes every generated call through the
hand-written `request` helper in `src/api/client.ts`, so error shape handling stays
in one place.

## Checks

```sh
cd backend  && make check        # gofmt + go vet + go test (needs Postgres)
cd frontend && pnpm check        # tsc + oxlint
cd frontend && pnpm build        # typecheck + production build
```

The backend tests cover the rules rather than the plumbing: sequence enforcement,
WIP limits, capacity refusal without partial consumption, consumption recorded once
at start, cancellation, and refusing to delete a resource still in use. They use
`TEST_DATABASE_URL` (default: a local `code_interview_test`) and skip when no
database is reachable.

## A first walkthrough

1. **Setup, Resources** and create a material with a capacity, for example
   `Steel, material, kg, 100`.
2. **Setup, Processes** and add a step that consumes it, then a second step so
   there is a sequence.
3. **Orders, New order**; open it and **Start** the first step. Consumption appears
   immediately. **Complete** it, then start the next.
4. **Board** shows the order in the right column, the counters, and the resource
   consumption accumulating. Completing every step marks the order done.
