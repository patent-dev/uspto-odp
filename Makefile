.PHONY: generate test test-integration lint fmt coverage tidy examples

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

lint:
	golangci-lint run

fmt:
	gofmt -w .

coverage:
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -n 1

tidy:
	go mod tidy

# examples runs the demo against the live API to refresh demo/examples (recorded
# request/response pairs). Requires the API credentials in the environment.
# The weekly response-watch workflow runs this and diffs the result.
examples:
	cd demo && go run .
