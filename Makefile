.PHONY: build run dev test lint bench clean

default: build

build:
	@printf "Building the application...\n"
	@pnpm build
	@go build -ldflags="-s -w" -trimpath -o app

run: build
	@printf "Running the binary...\n"
	@./app

dev:
	@air

test:
	@printf "Starting tests...\n"
	@./scripts/test.sh

lint:
	@printf "Linting files...\n"
	@pnpm lint

bench:
	@printf "Starting benchmark...\n"
	@./scripts/bench.sh

clean:
	@printf "Cleaning cache and build artifacts...\n"
	@rm -f app
	@rm -f coverage.out
	@rm -rf tmp/
	@go clean -testcache
	@go clean -modcache
