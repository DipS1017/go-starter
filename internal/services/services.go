package services

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/webpoint-solutions-llc/dba/internal/config"
	"github.com/webpoint-solutions-llc/dba/internal/db"
	"github.com/webpoint-solutions-llc/dba/internal/db/sqlc"
	"github.com/webpoint-solutions-llc/dba/internal/pkg/apple"
	"github.com/webpoint-solutions-llc/dba/internal/pkg/redisclient"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type S3Client struct {
	client    *s3.Client
	presigner *s3.PresignClient
}

type Service struct {
	q      *sqlc.Queries
	redis  *redis.Client
	db     *pgxpool.Pool
	logger *slog.Logger

	s3Client *S3Client

	appleClient *apple.AppleConfig
	esClient    *elasticsearch.Client
}

func NewService() *Service {
	s3Client, err := newS3Client(context.Background())
	if err != nil {
		slog.Error("failed to create S3 client", "error", err)
		return nil
	}

	appleClient := apple.NewAppleConfig(config.Cfg.AppleTeamID, config.Cfg.AppleClientID, config.Cfg.AppleRedirectURI, config.Cfg.AppleKeyID)

	if err := appleClient.LoadPrivateKeyFromFile(config.Cfg.AppleCertificatePath); err != nil {
		slog.Error("Failed to load private key from file", "error", err)
	}

	s := &Service{
		q:           db.SqlcQuery,
		redis:       redisclient.Client,
		logger:      slog.Default().With("component", "services"),
		s3Client:    s3Client,
		db:          db.Client,
		appleClient: appleClient,
	}
	return s
}
