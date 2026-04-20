package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(dsn string) error {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db for migrations: %w", err)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(migrations)

	err = goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	err = goose.Up(sqlDB, "migrations")
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
