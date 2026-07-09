DB_URL ?= postgres://carbridge:carbridge@localhost:5432/carbridge?sslmode=disable
MIGRATE = migrate -path db/migrations -database "$(DB_URL)"

PROTO_FILES = $(shell find api/proto -name '*.proto')

.PHONY: run build tidy test vet fmt proto proto-tools db-up db-down migrate-up migrate-down migrate-create

run:
	go run ./cmd/web

# Regenerate gRPC stubs from api/proto into internal/delivery/grpc/gen.
# Requires protoc + the Go plugins (see: make proto-tools).
proto:
	protoc \
		--proto_path=api/proto \
		--go_out=. --go_opt=module=car-bridge \
		--go-grpc_out=. --go-grpc_opt=module=car-bridge \
		$(PROTO_FILES)

# Install the protoc Go plugins (protoc itself: `brew install protobuf`).
proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

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
