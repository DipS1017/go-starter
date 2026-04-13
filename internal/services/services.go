package services

import (
	"context"
	"log/slog"

	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
	"github.com/webpoint-solutions-llc/go-starter/internal/db"
	"github.com/webpoint-solutions-llc/go-starter/internal/db/sqlc"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
	"github.com/webpoint-solutions-llc/go-starter/internal/pkg/apple"
	"github.com/webpoint-solutions-llc/go-starter/internal/pkg/redisclient"
)

// querier extends the sqlc-generated Querier with transaction support.
// *sqlc.Queries satisfies this interface.
type querier interface {
	sqlc.Querier
	WithTx(tx pgx.Tx) *sqlc.Queries
}

// cache is the minimal Redis interface used by the service layer.
// *redis.Client satisfies this interface.
type cache interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	TxPipeline() redis.Pipeliner
}

// storage is the interface for object storage operations.
// *S3Client satisfies this interface.
type storage interface {
	UploadFile(ctx context.Context, params dto.S3UploadParams) (dto.S3FileUpload, error)
	S3MediaURL(ctx context.Context, key string) (string, error)
	DeleteObjects(ctx context.Context, objects []s3types.ObjectIdentifier, bypassGovernance bool) error
}

type Service struct {
	q     querier
	cache cache
	store storage
	db    *pgxpool.Pool

	logger      *slog.Logger
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

	return &Service{
		q:           db.SqlcQuery,       // *sqlc.Queries satisfies querier
		cache:       redisclient.Client, // *redis.Client satisfies cache
		store:       s3Client,           // *S3Client satisfies storage
		db:          db.Client,
		logger:      slog.Default().With("component", "services"),
		appleClient: appleClient,
	}
}
