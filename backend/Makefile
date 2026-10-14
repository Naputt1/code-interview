GO ?= go
SWAG_VERSION := v1.16.4
SWAGGER2OPENAPI := swagger2openapi@7.0.8

.PHONY: run build test test-race vet lint fmt fmt-check check swagger

run:
	$(GO) run ./cmd/api

build:
	$(GO) build -o bin/api ./cmd/api

# Rule-level tests. They need a Postgres database; point TEST_DATABASE_URL at one
# (defaults to a local `code_interview_test`). Without a database they skip.
test:
	$(GO) test ./...

# Same tests under the race detector. This is what CI runs.
test-race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w .

# Report unformatted files without rewriting them, so `check` never edits the
# working tree. CI runs the same test.
fmt-check:
	@test -z "$$(gofmt -l .)" || { echo "gofmt: files need formatting:"; gofmt -l .; exit 1; }

# Needs golangci-lint on PATH. CI pins the version.
lint:
	golangci-lint run ./...

check: fmt-check vet lint test

# Regenerate the committed OpenAPI contract from the Go annotations in the
# handlers. The frontend's orval config reads api/swagger/openapi3.yaml.
swagger:
	$(GO) run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init \
		-g cmd/api/main.go -o api/swagger --outputTypes yaml,json
	npx --yes $(SWAGGER2OPENAPI) api/swagger/swagger.yaml -o api/swagger/openapi3.yaml
