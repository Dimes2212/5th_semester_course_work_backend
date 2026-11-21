POSTGRES_DB ?= campuslost
POSTGRES_USER ?= campuslost
MIGRATION_DATABASE_URL ?= postgres://campuslost:campuslost@postgres:5432/campuslost?sslmode=disable
BACKUP_DIR ?= backups
BACKUP_FILE ?=

.PHONY: db-up db-down db-logs db-shell db-tables db-backup db-restore migrate-up migrate-down test fmt vet

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

db-backup:
	mkdir -p "$(BACKUP_DIR)"
	docker compose exec -T postgres pg_dump -U "$(POSTGRES_USER)" -d "$(POSTGRES_DB)" --format=custom > "$(BACKUP_DIR)/campuslost_$$(date +%Y%m%d_%H%M%S).dump"

db-restore:
	@test -n "$(BACKUP_FILE)" || (echo "BACKUP_FILE is required" && exit 1)
	@test -f "$(BACKUP_FILE)" || (echo "backup file not found: $(BACKUP_FILE)" && exit 1)
	docker compose exec -T postgres pg_restore -U "$(POSTGRES_USER)" -d "$(POSTGRES_DB)" --clean --if-exists < "$(BACKUP_FILE)"

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
