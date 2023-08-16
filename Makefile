build:
	@echo "Building the project..."
	go build -o steelcut
	@echo "Build complete!"

docs:
	@echo "Generating documentation..."
	godoc -http=:6060
	@echo "Documentation server started!"

test:
	@echo "Running tests..."
	go test -v ./...
	@echo "Tests complete!"

cover:
	@echo "Measuring test coverage..."
	go test -cover ./...
	@echo "Coverage complete!"

run:
	@echo "Running the project..."
	go run main.go
	@echo "Execution complete!"

clean:
	@echo "Cleaning up..."
	rm -f steelcut/steelcut
	@echo "Clean up complete!"

vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet complete!"

lint:
	@echo "Running linter..."
	golint ./...
	@echo "Linting complete!"

fmt:
	@echo "Running gofmt on all .go files..."
	find . -name '*.go' -exec gofmt -w {} \;
	@echo "Done formatting!"

mod:
	@echo "Tidying and verifying module dependencies..."
	go mod tidy
	go mod verify
	@echo "Dependencies are tidy and verified!"

.PHONY: fmt build test run clean vet lint mod
