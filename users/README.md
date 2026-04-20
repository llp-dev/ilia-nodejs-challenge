# Users

Users microservice built with Go, Gin, and PostgreSQL.

## Requirements

- Go 1.21+
- PostgreSQL
- Make

## Running locally

```bash
USERS_DSN="postgres://user:password@localhost:5432/users?sslmode=disable" \
USERS_PORT=3002 \
USERS_JWT_SECRET=ILIACHALLENGE \
USERS_JWT_INTERNAL_SECRET=ILIACHALLENGE_INTERNAL \
USERS_WALLET_URL=http://localhost:3001 \
make run
```

## Running with Docker

```bash
cp .env.example .env  # fill in the values
make up               # build and start all containers (wallet + users)
make down             # stop containers
make clean            # stop containers, remove volumes and images
```

## Building

```bash
make build
# binary output: bin/api
```

## Testing

```bash
make test
```

To check coverage:

```bash
go test ./... -cover
```

## Environment Variables

| Variable                      | Required | Default | Description                                                          |
|-------------------------------|----------|---------|----------------------------------------------------------------------|
| `USERS_DSN`                   | yes      | —       | PostgreSQL DSN                                                       |
| `USERS_PORT`                  | yes      | —       | Port the HTTP server listens on                                      |
| `USERS_JWT_SECRET`            | yes      | —       | HS256 secret for external JWT tokens (`ILIACHALLENGE`)               |
| `USERS_JWT_INTERNAL_SECRET`   | yes      | —       | HS256 secret for service-to-service calls (`ILIACHALLENGE_INTERNAL`) |
| `USERS_WALLET_URL`            | yes      | —       | Base URL of the Wallet service (e.g. `http://wallet-api:3001`)       |
| `USERS_RELEASE`               | no       | `false` | Set to `true` to run Gin in release mode (less logging)              |

## Migrations

Migrations run automatically on startup via goose. No separate step required.

## Endpoints

### Public

| Method | Path        | Description              |
|--------|-------------|--------------------------|
| GET    | `/health`   | Health check             |
| POST   | `/users`    | Register a new user      |
| POST   | `/sessions` | Login and receive a JWT  |

### Authenticated (requires `Authorization: Bearer <token>`)

| Method | Path         | Description                        |
|--------|--------------|------------------------------------|
| GET    | `/users/me`  | Get the authenticated user profile |
| PUT    | `/users/me`  | Update name, email, or password    |

### Internal (requires JWT signed with `ILIACHALLENGE_INTERNAL`)

| Method | Path          | Description                                      |
|--------|---------------|--------------------------------------------------|
| GET    | `/users/:id`  | Look up a user by ID — used by the Wallet service |

## API Reference

### POST /users

```json
{
  "name": "Alice",
  "email": "alice@example.com",
  "password": "password123"
}
```

Returns `201` with the created user. Returns `409` if the email is already taken.

### POST /sessions

```json
{
  "email": "alice@example.com",
  "password": "password123"
}
```

Returns `200` with `{ "token": "<jwt>", "user": { ... } }`.

### PUT /users/me

All fields are optional — only the provided fields are updated.

```json
{
  "name": "Alicia",
  "email": "alicia@example.com",
  "password": "newpassword123"
}
```

## Inter-service Communication

The Wallet service calls `GET /users/:id` to validate that the `user_id` in a wallet-creation request belongs to the authenticated user. This route is protected by the internal JWT secret (`ILIACHALLENGE_INTERNAL`).

Communication is REST over HTTP. In the Docker environment both services are on the same Docker network and the Wallet service reaches Users at `http://users-api:3002`.
