COMPOSE = docker compose

.PHONY: up down clean

up:
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down

clean:
	$(COMPOSE) down --volumes --rmi all
