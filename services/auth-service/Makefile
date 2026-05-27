run:
	go run ./cmd/main.go

build:
	go build -o bin/auth-service ./cmd/main.go

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

migrate-up:
	goose -dir internal/database/migrations postgres "host=$${DB_HOST} port=$${DB_PORT} user=$${DB_USER} password=$${DB_PASSWORD} dbname=$${DB_NAME} sslmode=disable" up

migrate-down:
	goose -dir internal/database/migrations postgres "host=$${DB_HOST} port=$${DB_PORT} user=$${DB_USER} password=$${DB_PASSWORD} dbname=$${DB_NAME} sslmode=disable" down

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v