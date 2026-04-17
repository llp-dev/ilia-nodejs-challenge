# Wallet

Wallet microservice built with Go, Gin, and PostgreSQL.

## Requirements

- Go 1.21+
- PostgreSQL
- Make

## Running

```bash
WALLET_DSN="postgres://user:password@localhost:5432/wallet?sslmode=disable" make run
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

| Variable         | Default    | Description                                             |
|------------------|------------|---------------------------------------------------------|
| `WALLET_DSN`     | (required) | PostgreSQL DSN                                          |
| `WALLET_PORT`    | `3001`     | Port the HTTP server listens on                         |
| `WALLET_RELEASE` | `false`    | Set to `true` to run Gin in release mode (less logging) |

## Migrations

Migrations run automatically on startup via goose. No separate step required.

## Endpoints

| Method | Path                          | Description                    |
|--------|-------------------------------|--------------------------------|
| GET    | `/health`                     | Health check                   |
| GET    | `/wallets`                    | List all wallets               |
| GET    | `/wallets/:id`                | Get wallet by ID               |
| POST   | `/wallets`                    | Create a wallet                |
| PUT    | `/wallets/:id`                | Update wallet description      |
| POST   | `/wallets/:id/transactions`   | Post a transaction to a wallet |

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
