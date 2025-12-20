package utils

import (
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/constants"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
	"github.com/webpoint-solutions-llc/go-starter/internal/errorhandler"
)

// It returns true if the string is a valid UUID, false otherwise.
func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func StringToNullUUID(id string) (uuid.NullUUID, error) {
	if id == "" {
		return uuid.NullUUID{}, nil
	}

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return uuid.NullUUID{}, err
	}

	return uuid.NullUUID{
		UUID:  parsedUUID,
		Valid: true,
	}, nil
}

func ArrStringToArrUUID(arr []string) ([]uuid.UUID, error) {
	uuids := make([]uuid.UUID, 0, len(arr)) // optimized allocation
	for _, id := range arr {
		if id == "" {
			continue
		}
		parsedUUID, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, parsedUUID)
	}
	return uuids, nil
}

// UUIDsToString joins UUIDs with "__"
func UUIDsToString(uuids []uuid.UUID) string {
	strs := make([]string, len(uuids))
	for i, u := range uuids {
		strs[i] = u.String()
	}
	return strings.Join(strs, "__")
}

func StringToUUIDs(s string) ([]uuid.UUID, error) {
	parts := strings.Split(s, "__")
	uuids := make([]uuid.UUID, len(parts))
	for i, part := range parts {
		id, err := uuid.Parse(part)
		if err != nil {
			return nil, err
		}
		uuids[i] = id
	}
	return uuids, nil
}

// GetUserIDFromContext extracts the userID from JWT claims in echo.Context
// Returns (uuid.UUID, errorhandler.HttpError) for handler use
func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	claims, ok := c.Get("claims").(*dto.CustomClaims)
	if !ok || claims == nil {
		return uuid.UUID{}, errorhandler.ErrorBadRequest(constants.MsgReLogin)
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, errorhandler.ErrorBadRequest(constants.MsgReLogin)
	}
	return userID, nil
}
