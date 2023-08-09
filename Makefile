fmt:
	@echo "Running gofmt on all .go files..."
	@find . -name '*.go' -exec gofmt -w {} \;
	@echo "Done formatting!"

