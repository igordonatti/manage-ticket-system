.PHONY: up down logs ps reset frontend-build

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

reset:
	docker compose down -v

frontend-build:
	npm --prefix frontend run build

