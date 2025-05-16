build:
	docker compose build --no-cache

pull:
	docker compose pull

up:
	docker compose up -d --pull always
	make setup

restart_mm:
	docker compose restart marketmaker

down:
	docker compose down

ash_mm:
	docker compose exec marketmaker ash

ash_ledgie:
	docker compose exec ledgie ash

bash:
	docker compose exec oms bash

ps:
	docker compose ps

logs:
	docker compose logs -f $(s)

logs_error:
	docker compose logs -f $(s) | grep --color=auto --line-buffered -i -e 'error' -

logs_mm:
	docker compose logs -f marketmaker marketmaker-dbwriter

logs_ledgie:
	docker compose logs -f ledgie

setup:
	docker compose exec oms sh -c 'until nc -z postgres 5432; do sleep 1; done'
	docker compose exec postgres sh /setup/setup-postgres.sh
	docker compose exec oms sh -c 'until nc -z redpanda-1 9092; do sleep 1; done'
	docker compose exec redpanda-1 sh /setup/setup-redpanda.sh
	docker compose exec oms sh -c 'until nc -z marketmaker 8080; do echo "waiting for marketmaker..."; sleep 1; done'
	docker compose exec marketcache sh /setup/setup-marketcache.sh
	docker compose exec oms sh /setup/setup-oms.sh
	@echo setup complete

test:
	docker compose exec marketmaker go test --timeout 100s --count 1 --cover $(TEST_OPTS) ./...

integration-test:
	docker compose exec oms sh -c 'until nc -z marketmaker 8080; do echo "waiting for marketmaker..."; sleep 1; done'
	docker compose exec marketmaker go test -v -tags integration -count=1 --timeout 300s $(TEST_OPTS) ./integration/...

.PHONY: all build pull up down bash ps logs setup