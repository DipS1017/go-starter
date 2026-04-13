package services

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
	"github.com/webpoint-solutions-llc/go-starter/internal/constants"
	"github.com/webpoint-solutions-llc/go-starter/internal/errorhandler"
	"github.com/webpoint-solutions-llc/go-starter/internal/types"
)

func (s *Service) checkForWrongPasswordAttempt(ctx context.Context, key string) error {
	wrongAttempts, err := s.cache.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		s.logger.Error("Redis Get error "+key, "err", err)
		return err
	}

	var attempt int64
	if err == redis.Nil {
		// If the key doesn't exist, start the count from 0
		attempt = 0
	} else {
		attempt, err = strconv.ParseInt(wrongAttempts, 10, 64)
		if err != nil {
			s.logger.Error("Parse Error", "err", err)
			return nil
		}
	}

	if attempt >= config.Cfg.MaxWrongPasswordAttempt {
		return errorhandler.ErrorBadRequest(constants.MsgMaxWrongPasswordAttempt)
	}

	// Increment the counter and set an expiration of 60 seconds
	pipe := s.cache.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 60*time.Second)
	if _, err = pipe.Exec(ctx); err != nil {
		s.logger.Error("Redis Exec error", "err", err)
	}
	return nil
}

func extractNameFromClaims(claim types.AppleIDTokenClaims) string {
	var name string

	// Priority 1: Full name
	if claim.Name != "" {
		name = claim.Name
	} else if claim.GivenName != "" || claim.FamilyName != "" {
		// Priority 2: Constructed from first and last name
		name = strings.TrimSpace(claim.GivenName + " " + claim.FamilyName)
	} else if claim.Email != "" {
		// Priority 3: Use email prefix
		parts := strings.Split(claim.Email, "@")
		if len(parts) > 0 {
			name = parts[0]
		}
	}
	return name
}
