# How AI tools were used

This project was built with an AI coding agent, OpenCode, working under my
direction. I made the decisions, the agent wrote and ran code, and I verified the
result by running it.

## What I decided

- The domain model and the rules in section 2 of `docs/summary.md`: steps and
  resources as data, the sequence as a `position`, consumption as an append-only
  ledger, and consumption written at step *start*.
- The stack, chosen to match tooling I already work with: Go, Gin, GORM and
  Postgres on one side; React, TanStack Router and Query, Orval, shadcn and Zustand
  on the other. Also the arrangement where swaggo annotations generate the
  committed OpenAPI spec and Orval generates the client from it.
- The scope cuts: no authentication, no streaming, no containerised deployment.

## What the agent did

- Scaffolded both apps and the dev orchestration (`fog.json`,
  `infra/docker-compose.yml`), and wrote the first draft of the Go handlers,
  service, migrations and React screens from the model above.
- Wrote the seed data and the service-level tests from the rules I specified.
- Ran the loop that caught the mistakes: `go build`, `go vet` and `gofmt`,
  `make swagger`, `pnpm generate`, `tsc` and `vite build`, then the app in a
  browser against a real PostgreSQL, calling the API with `curl` where a UI check
  was enough.

## Corrections that came out of running it

- Every create form silently sent an empty request body. Orval's generated client
  passes an already-stringified `body`, while the hand-written mutator read a
  `data` field and overwrote `body` with `undefined`. Nothing caught it. The code
  typechecked, the build succeeded and every `GET` screen looked right. It only
  surfaced when the API answered `400 invalid request body: EOF` in the browser.
- The UI claimed consumption was recorded when a step *completed*, while the code
  recorded it on *start*. The browser showed a running step already carrying
  consumption, 1.0 hour of welding station, which is what exposed the mismatch. I
  kept the behaviour and fixed the copy.
- Order detail pages showed "Requires no resources". The order queries preloaded
  `steps.process` but not the nested recipe, so the UI was quietly missing data.
- Two dialog components initialised their form state in an effect. Unmounting the
  closed dialog and initialising from props removed the redundant state sync.

## What I verified myself

The rules are asserted in `backend/internal/service/service_test.go` and pass
against Postgres. I walked the flows in a browser: creating resources and
processes, creating an order, starting and completing steps, watching consumption
accumulate, and hitting the refusals for a full WIP limit, exhausted capacity and
cancelling an order that has already started.

Anything the agent asserted that I did not run, I treated as unverified.
