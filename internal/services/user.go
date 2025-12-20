package services

import (
	"context"
	"errors"

	goaway "github.com/TwiN/go-away"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/webpoint-solutions-llc/go-starter/internal/constants"
	"github.com/webpoint-solutions-llc/go-starter/internal/db/sqlc"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
	"github.com/webpoint-solutions-llc/go-starter/internal/errorhandler"
	"github.com/webpoint-solutions-llc/go-starter/internal/pkg/redisclient"
)

func (s *Service) UpdateProfileImage(ctx context.Context, userID uuid.UUID, imagePath string) (sqlc.UpdateProfileImageRow, error) {
	return s.q.UpdateProfileImage(ctx, sqlc.UpdateProfileImageParams{
		ID:           userID,
		ProfileImage: &imagePath,
	})
}

func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (sqlc.GetUserByIdRow, error) {
	return s.q.GetUserById(ctx, userID)
}

func (s *Service) UpdateUser(ctx context.Context, userID uuid.UUID, userInfo dto.UpdateUserInfoParams) error {
	tx, txErr := s.db.Begin(ctx)

	if txErr != nil {
		return errors.New("cannot start transaction: " + txErr.Error())
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if txErr != nil {
			tx.Rollback(ctx)
		}
	}()

	q := s.q.WithTx(tx)

	var updateErr error

	// Profanity check for User fields
	if userInfo.User.Name != nil && goaway.IsProfane(*userInfo.User.Name) {
		return errorhandler.ErrorBadRequest(constants.MsgProfanityViolation)
	}
	if userInfo.User.Username != nil && goaway.IsProfane(*userInfo.User.Username) {
		return errorhandler.ErrorBadRequest(constants.MsgProfanityViolation)
	}

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

func (s *Service) AddActiveUser(ctx context.Context, userId string) error {
	return redisclient.AddActiveUser(ctx, userId)
}

func (s *Service) RemoveActiveUser(ctx context.Context, userId string) error {
	return redisclient.RemoveActiveUser(ctx, userId)
}

func (s *Service) GetActiveUsers(ctx context.Context) ([]string, error) {
	return redisclient.GetActiveUsers(ctx)
}

func (s *Service) IsActiveUser(ctx context.Context, userId string) bool {
	return redisclient.GetIsUserActive(ctx, userId)
}

func (s *Service) ClearAllActiveUser(ctx context.Context) error {
	return redisclient.ClearAllActiveUsers(ctx)
}

// extractLocationStrings extracts string values from optional location pointers
func extractLocationStrings(city, state, country *string) (string, string, string) {
	var cityStr, stateStr, countryStr string
	if city != nil {
		cityStr = *city
	}
	if state != nil {
		stateStr = *state
	}
	if country != nil {
		countryStr = *country
	}
	return cityStr, stateStr, countryStr
}
