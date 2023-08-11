build:
	@echo "Building the project..."
	@go build -o steelcut
	@echo "Build complete!"

test:
	@echo "Running tests..."
	@go test ./...
	@echo "Tests complete!"

run:
	@echo "Running the project..."
	@go run main.go
	@echo "Execution complete!"

clean:
	@echo "Cleaning up..."
	@rm steelcut/steelcut
	@echo "Clean up complete!"

vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete!"

fmt:
	@echo "Running gofmt on all .go files..."
	@find . -name '*.go' -exec gofmt -w {} \;
	@echo "Done formatting!"

.PHONY: fmt build test run clean vet lint
