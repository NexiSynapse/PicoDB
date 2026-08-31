build:
	go build -o picodb ./cmd/picodb

test:
	go test ./...

integration:
	go test -tags=integration ./...

vet:
	go vet ./...

fmt-check:
	test -z "$$(gofmt -l .)"

deps:
	GOPROXY=off GOTOOLCHAIN=local go list -m all
	GOPROXY=off GOTOOLCHAIN=local go mod verify

demo-crash:
	./scripts/crash_demo.sh

check: fmt-check vet test integration
