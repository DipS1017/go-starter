package db

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var TestDBURL string

func mustStartPostgresContainer() (func(context.Context) error, error) {
	var (
		dbName = "database"
		dbPwd  = "password"
		dbUser = "user"
	)

	dbContainer, err := postgres.Run(
		context.Background(),
		"postgres:latest",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPwd),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	dbhost, err := dbContainer.Host(context.Background())
	if err != nil {
		return func(ctx context.Context) error {
			return dbContainer.Terminate(ctx)
		}, err
	}

	dbport, err := dbContainer.MappedPort(context.Background(), "5432/tcp")
	if err != nil {
		return func(ctx context.Context) error {
			return dbContainer.Terminate(ctx)
		}, err
	}

	TestDBURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPwd, dbhost, dbport.Port(), dbName)

	return func(ctx context.Context) error {
		return dbContainer.Terminate(ctx)
	}, err
}

func TestMain(m *testing.M) {
	teardown, err := mustStartPostgresContainer()
	if err != nil {
		log.Fatalf("could not start postgres container: %v", err)
	}

	m.Run()

	if teardown != nil && teardown(context.Background()) != nil {
		log.Fatalf("could not teardown postgres container: %v", err)
	}
}

func TestNew(t *testing.T) {
	srv := OpenDbConnection(TestDBURL)
	if srv == nil {
		t.Fatal("OpenDbConnection() returned nil")
	}
}
