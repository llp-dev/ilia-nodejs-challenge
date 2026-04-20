COMPOSE = docker compose

.PHONY: up down clean test k6

up:
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down

clean:
	$(COMPOSE) down --volumes --rmi all

test:
	cd users && go vet ./... && go test ./...
	cd wallet && go vet ./... && go test ./...

k6:
	k6 run k6/users_sequence.js
	k6 run k6/wallet_sequence.js
