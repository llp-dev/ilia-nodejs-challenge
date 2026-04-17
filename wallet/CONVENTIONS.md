# Conventions

Technical decisions and rationale for the wallet service.

---

## IDs

All entities use UUID v4 as primary key, generated in the repository layer before the INSERT. This avoids relying on DB-generated serials and lets the service own the ID without a round-trip.

`user_id` on a wallet is an opaque UUID from an external user service. It is stored as-is — the wallet service does not validate its existence or format beyond being a non-empty string.

## Monetary values

`float64` is not used for money — binary floating-point arithmetic is not exact (e.g. `0.1 + 0.2 ≠ 0.3`). `int64` cents was also considered but carries an implicit unit convention that is invisible at API and DB boundaries.

**Decision**: `github.com/shopspring/decimal` everywhere — models, repositories, API payloads. Stored as `NUMERIC` in PostgreSQL.

## Transaction write path

Balance check is embedded directly in the UPDATE predicate instead of a prior `SELECT … FOR UPDATE`:

```sql
UPDATE wallets
SET balance = balance + $1, updated_at = $2
WHERE id = $3 AND (balance + $1) >= 0
```

`RowsAffected == 0` means the condition was not met (insufficient balance or wallet not found). This is safe under concurrent writes because a PostgreSQL row UPDATE is atomic. A DB transaction wraps the balance UPDATE and the transaction INSERT together — a `defer tx.Rollback()` ensures rollback on any failure before commit.

Positive values credit the wallet, negative values debit it.

## Balance constraint

`CHECK (balance >= 0)` is on the `wallets` table as a DB-level guard. The conditional UPDATE in the repository is the primary enforcement; the constraint is a last-resort safety net.

## Database driver

`pgxpool` (`github.com/jackc/pgx/v5`) is used instead of `database/sql` + a driver shim. It provides native PostgreSQL protocol support, built-in connection pooling, and better context propagation. A single pool is created at startup and shared across all repositories.

## Configuration

`WALLET_DSN` is the only required env var. The app fails fast on startup if it is absent — there is no default credential. All other vars have safe defaults.

## Migrations

`goose` SQL files live in `internal/db/migrations/` and are embedded in the binary via `//go:embed`. `goose.Up` runs on startup before the HTTP server starts — deployments are self-migrating with no separate job.

## Server struct

`server.Server` owns both the `*pgxpool.Pool` and the `*gin.Engine`. Repositories and handlers are wired inside `setupRoutes()`. This keeps `main.go` limited to process lifecycle (config, connect, migrate, run) and makes dependencies explicit.
