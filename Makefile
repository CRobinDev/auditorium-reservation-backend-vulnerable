# Makefile

# Use the .env file to load environment variables
-include .env

# Default environment variables for Docker Compose
POSTGRES_URL := postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

MIGRATE_CMD=docker-compose run --rm migrate -database "postgres://auditorium-reservation-backend:bebek123@localhost:5432/auditorium-reservation-backend?sslmode=disable" -path migration/
SEED_CMD=docker compose exec -T db psql -U $(DB_USER) -d $(DB_NAME) -W $(DB_PASS) < database/seeder/

# Targets for different migration commands
.PHONY: up down redo status version force seed-up seed-down run stop build-push

# Apply all migrations
migrate-up:
	$(MIGRATE_CMD) up

# Rollback the most recent migration
migrate-down:
	$(MIGRATE_CMD) down

# Revert the most recent migration and then reapply it
migrate-redo:
	$(MIGRATE_CMD) redo

# Show the status of migrations
migrate-status:
	$(MIGRATE_CMD) status

# Show the current version of the database
migrate-version:
	$(MIGRATE_CMD) version

# Force a specific version (replace <version> with the desired version)
migrate-force:
	$(MIGRATE_CMD) force $(version)

# Seeding commands
seed-up:
	$(SEED_CMD)$(word 2,$(MAKECMDGOALS)).up.sql

seed-down:
	$(SEED_CMD)$(word 2,$(MAKECMDGOALS)).down.sql

run:
	docker compose up -d
	make migrate-up

stop:
	docker compose down

build:
	docker compose build

# Catch the environment argument
%:
	@:
