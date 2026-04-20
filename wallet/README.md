# Wallet

Wallet microservice built with Go, Gin, and PostgreSQL.

## Requirements

- Go 1.21+
- PostgreSQL
- Make

## Running locally

```bash
WALLET_DSN="postgres://user:password@localhost:5432/wallet?sslmode=disable" \
WALLET_PORT=3001 \
WALLET_JWT_SECRET=ILIACHALLENGE \
make run
```

## Running with Docker

```bash
cp .env.example .env  # fill in the values
make up               # build and start all containers
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

## Formatting

```bash
make fmt
```

## Environment Variables

| Variable            | Required | Default        | Description                                             |
|---------------------|----------|----------------|---------------------------------------------------------|
| `WALLET_DSN`        | yes      | —              | PostgreSQL DSN                                          |
| `WALLET_PORT`       | yes      | —              | Port the HTTP server listens on                         |
| `WALLET_JWT_SECRET` | yes      | —              | HS256 secret used to validate JWT tokens                |
| `WALLET_RELEASE`    | no       | `false`        | Set to `true` to run Gin in release mode (less logging) |

## Migrations

Migrations run automatically on startup via goose. No separate step required.

## Endpoints

All endpoints except `/health` require a valid HS256 JWT token in the `Authorization: Bearer <token>` header.

| Method | Path                          | Auth | Description                    |
|--------|-------------------------------|------|--------------------------------|
| GET    | `/health`                     | —    | Health check                   |
| GET    | `/wallets`                    | ✓    | List all wallets               |
| GET    | `/wallets/:id`                | ✓    | Get wallet by ID               |
| POST   | `/wallets`                    | ✓    | Create a wallet                |
| PUT    | `/wallets/:id`                | ✓    | Update wallet description      |
| POST   | `/wallets/:id/transactions`   | ✓    | Post a transaction to a wallet |

### POST /wallets

```json
{
  "user_id": "uuid",
  "description": "my wallet"
}
```

### PUT /wallets/:id

```json
{
  "description": "updated description"
}
```

### POST /wallets/:id/transactions

```json
{
  "value": "10.50",
  "description": "purchase",
  "operation_id": "uuid"
}
```

Positive `value` credits the wallet, negative `value` debits it.
