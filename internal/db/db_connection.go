package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
	"github.com/webpoint-solutions-llc/go-starter/internal/db/sqlc"
)

var (
	Client    *pgxpool.Pool
	SqlcQuery *sqlc.Queries
)

func OpenDbConnection(connStr string) *pgxpool.Pool {
	// Reuse Connection
	if Client != nil {
		return Client
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbInstance, err := pgxpool.New(ctx, config.Cfg.PostgresqlURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	dbInstance.Config().MaxConns = 10
	dbInstance.Config().MaxConnIdleTime = 30 * time.Second

	Client = dbInstance
	SqlcQuery = sqlc.New(Client)
	return dbInstance
}

func CloseDbConnection() {
	if Client != nil {
		Client.Close()
	}
}
