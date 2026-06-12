.PHONY: generate test test-integration check-integration refresh-fixtures lint fmt coverage tidy

# generate re-applies the OpenAPI fixes (if a script is present) and regenerates
# the typed client via the //go:generate directives.
generate:
	@if [ -x scripts/fix-openapi.sh ]; then ./scripts/fix-openapi.sh; \
	elif [ -x scripts/convert-openapi.sh ]; then ./scripts/convert-openapi.sh; fi
	go generate ./...

test:
	go test -race -count=1 ./...

test-integration:
	go test -race -count=1 -tags=integration ./...

# check-integration verifies every exported endpoint Client method has a
# per-endpoint TestIntegration<Method> in a //go:build integration test file.
check-integration:
	./scripts/check-integration-coverage.sh

lint:
	golangci-lint run

fmt:
	gofmt -w .

coverage:
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -n 1

tidy:
	go mod tidy

# refresh-fixtures deliberately re-captures the committed golden response bodies
# under testdata/strictdecode/ from the live API (requires USPTO_API_KEY). The
# deterministic tests (TestFixtures in decode_examples_test.go) read the COMMITTED
# copies, so re-capturing does not change test behavior until the new bodies are
# reviewed and committed. Search/list goldens are truncated by hand afterward.
refresh-fixtures:
	./scripts/refresh-fixtures.sh
	@echo "testdata/strictdecode updated; review 'git diff testdata/strictdecode' before committing."
