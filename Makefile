.PHONY: build run

default: build

build:
	@printf "Building the application...\n"
	@go build -ldflags="-s -w" -o app

run: build
	@printf "Running the binary...\n"
	@./app
