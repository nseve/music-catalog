include .env
export

MIGRATE = docker compose run --rm mc-postgres-migrate
DB_URL 	= postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@mc-postgres:5432/${POSTGRES_DB}?sslmode=disable

db-up:
	@docker compose up -d mc-postgres

db-down:
	@docker compose down mc-postgres

db-cleanup:
	@read -p "Очистить volume pgdata? Опасность утери данных. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down mc-postgres -v && \
		echo "Volume pgdata удалён."; \
	else \
		echo "Отменено. Данные не тронуты."; \
	fi

psql:
	@docker exec -it -e PGOPTIONS='-c search_path=mc' mc-postgres psql -U ${POSTGRES_USER} -d ${POSTGRES_DB}

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Отсутствует необходимый параметр name. Пример: make migrate-create name=init"; \
		exit 1; \
	fi; \
	$(MIGRATE) create -ext sql -dir //migrations -seq "$(name)"

migrate-up:
	@$(MIGRATE) -path //migrations -database "$(DB_URL)" up $(steps)

migrate-down:
	@$(MIGRATE) -path //migrations -database "$(DB_URL)" down $(steps)

migrate-version:
	@$(MIGRATE) -path //migrations -database "$(DB_URL)" version

migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "Отсутствует необходимый параметр version. Пример: make migrate-force version=N"; \
		exit 1; \
	fi; \
	$(MIGRATE) -path //migrations -database "$(DB_URL)" force $(version)