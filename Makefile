build:
	@go build -asan -o bin/main ./cmd/server

run: build
	@./bin/main
