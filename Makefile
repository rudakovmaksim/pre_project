local: lint test down up logs

lint:
	@echo "Starting linters..."
	golangci-lint run ./...

test:
	@echo "Starting test..."
	go test ./... -v

down:
	@echo "Stopping and clearing all old containers..."
	docker-compose down

up:
	@echo "Starting all containers..."
	docker-compose up -d --build

logs:
	@echo "Display of logs for Service ..."
	docker-compose logs app