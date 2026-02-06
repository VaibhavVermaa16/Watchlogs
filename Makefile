APP_NAME=watchlogs
GO=go
CMD_DIR=./cmd/server

.PHONY: help setup run build test fmt clean

help:
	@echo "Targets:"
	@echo "  setup  - download Go modules"
	@echo "  run    - run the server"
	@echo "  build  - build the server binary"
	@echo "  test   - run tests"
	@echo "  fmt    - format Go code"
	@echo "  clean  - remove build artifacts"

setup:
	$(GO) mod download

run:
	$(GO) run $(CMD_DIR)

build:
	$(GO) build -o bin/$(APP_NAME) $(CMD_DIR)

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

clean:
	rm -rf bin
