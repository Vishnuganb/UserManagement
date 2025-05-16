build:
	docker compose build --no-cache

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

logs:
	docker compose logs -f $(s)

test:
	docker compose exec usermanagement go test --timeout 100s --count 1 --cover $(TEST_OPTS) ./...

integration-test:
	docker compose exec usermanagement sh -c 'until nc -z kafka 9092; do echo "waiting for kafka..."; sleep 1; done'
	docker compose exec usermanagement sh -c 'until nc -z postgres 5432; do echo "waiting for postgres..."; sleep 1; done'
	docker compose exec usermanagement go test -v -tags integration -count=1 --timeout 300s $(TEST_OPTS) ./integration/...

.PHONY: build pull up down bash ps logs test integration-test