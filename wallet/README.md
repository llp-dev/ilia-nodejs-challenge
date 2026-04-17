# Wallet

Wallet microservice built with Go and Gin.

## Requirements

- Go 1.21+
- Make

## Running

```bash
make run
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

| Variable         | Default | Description                                              |
|------------------|---------|----------------------------------------------------------|
| `WALLET_PORT`    | `3001`  | Port the HTTP server listens on                          |
| `WALLET_RELEASE` | `false` | Set to `true` to run Gin in release mode (less logging)  |

## Endpoints

| Method | Path      | Description        |
|--------|-----------|--------------------|
| GET    | `/health` | Health check       |
