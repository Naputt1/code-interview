GO ?= go
SWAG_VERSION := v1.16.4
SWAGGER2OPENAPI := swagger2openapi@7.0.8

.PHONY: run build test vet fmt check swagger

run:
	$(GO) run ./cmd/api

build:
	$(GO) build -o bin/api ./cmd/api

# Rule-level tests. They need a Postgres database; point TEST_DATABASE_URL at one
# (defaults to a local `code_interview_test`). Without a database they skip.
test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w .

check: fmt vet test

# Regenerate the committed OpenAPI contract from the Go annotations in the
# handlers. The frontend's orval config reads api/swagger/openapi3.yaml.
swagger:
	$(GO) run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init \
		-g cmd/api/main.go -o api/swagger --outputTypes yaml,json
	npx --yes $(SWAGGER2OPENAPI) api/swagger/swagger.yaml -o api/swagger/openapi3.yaml
