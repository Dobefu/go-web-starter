.PHONY: build run test bench

default: build

build:
	@printf "Building the application...\n"
	@go build -ldflags="-s -w" -o app

run: build
	@printf "Running the binary...\n"
	@./app

test:
	@printf "Starting tests...\n"
	@./scripts/test.sh

bench:
	@printf "Starting benchmark...\n"
	@./scripts/bench.sh
