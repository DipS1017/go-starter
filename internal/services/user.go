package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/webpoint-solutions-llc/go-starter/internal/db/sqlc"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
	"github.com/webpoint-solutions-llc/go-starter/internal/pkg/redisclient"
)

func (s *Service) UpdateProfileImage(ctx context.Context, userID uuid.UUID, imagePath string) (sqlc.UpdateProfileImageRow, error) {
	return s.q.UpdateProfileImage(ctx, sqlc.UpdateProfileImageParams{
		ID:           userID,
		ProfileImage: &imagePath,
	})
}

func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (sqlc.GetUserByIDRow, error) {
	return s.q.GetUserByID(ctx, userID)
}

func (s *Service) UpdateUser(ctx context.Context, userID uuid.UUID, userInfo dto.UpdateUserInfoParams) error {
	tx, txErr := s.db.Begin(ctx)

	if txErr != nil {
		return errors.New("cannot start transaction: " + txErr.Error())
	}

	defer tx.Rollback(ctx)

	q := s.q.WithTx(tx)

	var updateErr error

	if !userInfo.User.IsEmpty() {
		params := sqlc.UpdateUserParams{ID: userID}
		copier.CopyWithOption(&params, &userInfo.User, copier.Option{IgnoreEmpty: true})

		updateErr = q.UpdateUser(ctx, params)
		if updateErr != nil {
			s.logger.Debug("Error on updating user: ", "err", updateErr.Error())
		}
	}

	tx.Commit(ctx)

	return updateErr
}

func (s *Service) AddActiveUser(ctx context.Context, userID string) error {
	return redisclient.AddActiveUser(ctx, userID)
}

func (s *Service) RemoveActiveUser(ctx context.Context, userID string) error {
	return redisclient.RemoveActiveUser(ctx, userID)
}

func (s *Service) GetActiveUsers(ctx context.Context) ([]string, error) {
	return redisclient.GetActiveUsers(ctx)
}

func (s *Service) IsActiveUser(ctx context.Context, userID string) bool {
	return redisclient.GetIsUserActive(ctx, userID)
}

func (s *Service) ClearAllActiveUser(ctx context.Context) error {
	return redisclient.ClearAllActiveUsers(ctx)
}
