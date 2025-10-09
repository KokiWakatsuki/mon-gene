init:
	docker compose down -v --remove-orphans
	docker system prune -a --volumes -f
	docker builder prune -a -f
	docker system df

build:
	docker compose build
	docker compose up -d

up:
	docker compose up -d

reup:
	docker compose down -v --remove-orphans
	docker system prune -a --volumes -f
	docker builder prune -a -f
	docker system df
	docker compose up -d

front-reup:
	docker compose restart front

prod-up:
	sudo docker compose -f docker-compose.stage.yml up -d