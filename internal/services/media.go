package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/webpoint-solutions-llc/go-starter/internal/db/sqlc"
)

func (s *Service) CreateMediaMeta(ctx context.Context, params sqlc.CreateMediaParams) (sqlc.CreateMediaRow, error) {
	return s.q.CreateMedia(ctx, params)
}

func (s *Service) GetUnusedMedia(ctx context.Context) ([]sqlc.GetUnusedMediaRow, error) {
	return s.q.GetUnusedMedia(ctx)
}

func (s *Service) DeleteMediaMetaByID(ctx context.Context, params uuid.UUID) (uuid.UUID, error) {
	return s.q.DeleteMediaByID(ctx, params)
}

func (s *Service) DeleteMediaMetaByIDBulk(ctx context.Context, params []uuid.UUID) error {
	return s.q.DeleteMediaByIDBulk(ctx, params)
}
