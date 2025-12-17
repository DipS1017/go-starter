package services

import (
	"context"

	"github.com/webpoint-solutions-llc/dba/internal/db/sqlc"
)

func (s *Service) GetSessionByRefreshTokenHash(ctx context.Context, hashedToken string) (sqlc.GetSessionByRefreshTokenHashRow, error) {
	return s.q.GetSessionByRefreshTokenHash(ctx, hashedToken)
}
