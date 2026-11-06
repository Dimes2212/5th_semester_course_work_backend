POSTGRES_DB ?= campuslost
POSTGRES_USER ?= campuslost
MIGRATION_DATABASE_URL ?= postgres://campuslost:campuslost@postgres:5432/campuslost?sslmode=disable

.PHONY: db-up db-down db-logs db-shell db-tables migrate-up migrate-down test fmt vet

db-up:
	docker compose up -d --wait postgres

db-down:
	docker compose down

db-logs:
	docker compose logs -f postgres

db-shell:
	docker compose exec postgres psql -U "$(POSTGRES_USER)" -d "$(POSTGRES_DB)"

db-tables:
	docker compose exec -T postgres psql -U "$(POSTGRES_USER)" -d "$(POSTGRES_DB)" -c '\dt'

migrate-up:
	docker compose run --rm migrate -path=/migrations -database="$(MIGRATION_DATABASE_URL)" up

migrate-down:
	docker compose run --rm migrate -path=/migrations -database="$(MIGRATION_DATABASE_URL)" down 1

test:
	go test ./...

fmt:
	gofmt -w $$(rg --files cmd internal -g '*.go')

vet:
	go vet ./...
