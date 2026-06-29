DB_URL ?= postgres://carbridge:carbridge@localhost:5432/carbridge?sslmode=disable
MIGRATE = migrate -path db/migrations -database "$(DB_URL)"

.PHONY: run build tidy test vet fmt db-up db-down migrate-up migrate-down migrate-create

run:
	go run ./cmd/web

build:
	go build ./...

tidy:
	go mod tidy

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

db-up:
	docker compose up -d

db-down:
	docker compose down

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

# usage: make migrate-create name=create_something
migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)
