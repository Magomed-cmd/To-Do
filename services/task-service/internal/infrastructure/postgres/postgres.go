package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"todoapp/services/task-service/internal/infrastructure/config"
)

func NewPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.PostgresURL())
	if err != nil {
		return nil, err
	}

	if cfg.Postgres.MaxConns > 0 {
		poolCfg.MaxConns = cfg.Postgres.MaxConns
	}

	if cfg.Postgres.MinConns > 0 {
		poolCfg.MinConns = cfg.Postgres.MinConns
	}

	if cfg.Postgres.MaxLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.Postgres.MaxLifetime
	}

	return pgxpool.NewWithConfig(ctx, poolCfg)
}
