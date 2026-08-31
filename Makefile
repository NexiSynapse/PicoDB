.PHONY: build test integration vet fmt-check fmt deps demo-crash check clean

BINARY := microdb
PKG    := ./...

build:
	go build -o $(BINARY) ./cmd/microdb

test:
	go test ./...

integration:
	go test -tags=integration ./...

vet:
	go vet ./...

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:"; gofmt -l .; exit 1)

fmt:
	gofmt -w .

deps:
	@echo "== go.mod ==" && cat go.mod
	@echo ""
	@echo "== go list (offline) ==" && GOPROXY=off GOTOOLCHAIN=local go list -m all
	@echo ""
	@echo "== go mod verify (offline) ==" && GOPROXY=off GOTOOLCHAIN=local go mod verify
	@echo ""
	@echo "== build (offline) ==" && GOPROXY=off GOTOOLCHAIN=local go build -o $(BINARY) ./cmd/microdb && echo "offline build OK" && rm -f $(BINARY)

demo-crash:
	./scripts/crash_demo.sh

# Full quality gate before recording (Plan 49)
check: fmt-check vet test integration

clean:
	rm -f $(BINARY) demo.wal *.wal
	go clean ./...
