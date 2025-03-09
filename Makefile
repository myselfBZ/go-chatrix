build:
	@go build -asan -o bin/migrate ./cmd/migrate/main.go
	@go build -asan -o bin/main ./cmd/server

run: build
	@#TODO thinkg of something better for migrations!!!
	@./bin/migrate ./cmd/migrate/schemas/00001_initial.sql
	@echo "migrated 00001_initial.sql successfully"
	@#
	@./bin/migrate ./cmd/migrate/schemas/00002_message_content_field.sql
	@echo "migrated 00002_message_content_field.sql successfully"
	@#
	@./bin/migrate ./cmd/migrate/schemas/00003_read_field.sql
	@echo "migrated 00003_read_field.sql successfully"
	@#
	@echo "Starting the server..."
	@./bin/main
