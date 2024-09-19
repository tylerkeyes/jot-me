# Simple Makefile for a Go project

# Build the application
all: build

build:
	@echo "Building..."
	
	
	@go build -o jot cmd/main.go

# Run the application
run:
	@go run cmd/main.go



# Test the application
test:
	@echo "Testing..."
	@go test ./... -v



# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f jot


.PHONY: all build run test clean
