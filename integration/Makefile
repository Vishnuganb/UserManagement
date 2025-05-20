pull:
	docker compose pull

up:
	docker compose up -d --pull always

restart:
	docker compose restart usermanagement

down:
	docker compose down

bash:
	docker compose exec usermanagement bash

ps:
	docker compose ps -a

db-tables:
	docker compose exec postgres psql -U root -d user_db -c '\dt'

createdb:
	docker compose exec postgres createdb --username=root --owner=root user_db

dropdb:
	docker compose exec postgres dropdb --if-exists user_db

migrateup:
	migrate -path ../internal/db/migration -database "postgresql://root:secret@localhost:5433/user_db?sslmode=disable" -verbose up

migratedown:
	migrate -path ../internal/db/migration -database "postgresql://root:secret@localhost:5433/user_db?sslmode=disable" -verbose down

# make logs s=usermanagement
logs:
	docker compose logs -f $(or $(s),usermanagement)

test:
	docker compose exec usermanagement go test --timeout 100s --count 1 --cover $(TEST_OPTS) ./...

integration-test:
	docker compose exec usermanagement sh -c 'until nc -z kafka 9092; do echo "waiting for kafka..."; sleep 1; done'
	docker compose exec usermanagement sh -c 'until nc -z postgres 5432; do echo "waiting for postgres..."; sleep 1; done'
	docker compose exec usermanagement go test -v -tags integration -count=1 --timeout 300s $(TEST_OPTS) ./integration/suite/...

.PHONY: pull up down bash ps logs test integration-test