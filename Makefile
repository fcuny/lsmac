.PHONY: build test test-cov vet fmt fmt-check ci clean

build:
	go build -o lsmac ./cmd/lsmac

test:
	go test -race ./...

test-cov:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "These files are not gofmt-clean:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

ci: fmt-check vet test build

clean:
	rm -f lsmac coverage.out
