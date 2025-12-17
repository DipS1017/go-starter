package redisclient

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

var Client *redis.Client

const (
	ActiveUserKey           = "active_users"
	PendingNotificationUser = "pending_notification_users"
)

var contentGroup singleflight.Group

func RedisConnect() error {
	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		return fmt.Errorf("REDIS_URL is not set")
	}
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")
	if dbStr == "" {
		dbStr = "0"
	}
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		return fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	Client = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}

// Close closes the Redis client.
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

func Set(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	_, err, _ := contentGroup.Do(key, func() (interface{}, error) {
		if found, err := Exists(ctx, key); err != nil {
			return nil, err
		} else if found {
			return nil, nil
		}

		_, err := Client.SetNX(ctx, key, value, exp).Result()
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func Get(ctx context.Context, key string) (string, bool, error) {
	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func Del(ctx context.Context, key string) error {
	if _, err := Client.Del(ctx, key).Result(); err != nil {
		return err
	}
	return nil
}

func Exists(ctx context.Context, key string) (bool, error) {
	n, err := Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// AddActiveUser adds a user to a Redis Set called `active_users`.
func AddActiveUser(ctx context.Context, user string) error {
	return Client.SAdd(ctx, ActiveUserKey, user).Err()
}

func AddPendingNotificationUser(ctx context.Context, user string) error {
	return Client.SAdd(ctx, PendingNotificationUser, user).Err()
}

func GetPendingNotificationUser(ctx context.Context) ([]string, error) {
	return Client.SMembers(ctx, PendingNotificationUser).Result()
}

func RemovePendingNotificationUser(ctx context.Context, user string) error {
	return Client.SRem(ctx, PendingNotificationUser, user).Err()
}

// RemoveActiveUser removes a user from the Set.
func RemoveActiveUser(ctx context.Context, user string) error {
	return Client.SRem(ctx, ActiveUserKey, user).Err()
}

// GetActiveUsers returns all members of the Set.
func GetActiveUsers(ctx context.Context) ([]string, error) {
	return Client.SMembers(ctx, ActiveUserKey).Result()
}

func GetIsUserActive(ctx context.Context, user string) bool {
	return Client.SIsMember(ctx, ActiveUserKey, user).Val()
}

func ClearAllActiveUsers(ctx context.Context) error {
	return Client.Del(ctx, ActiveUserKey).Err()
}

// AddNotificationToSet adds a value to a Redis set at the specified key.
// Parameters:
//
//	ctx - context for the operation
//	key - Redis key for the set
//	value - value to add to the set
//
// Returns:
//
//	error - error if the operation fails, nil otherwise
func AddNotificationToSet(ctx context.Context, key string, value string) error {
	return Client.SAdd(ctx, key, value).Err()
}

// AddNotificationToList pushes a value onto the left end of a Redis list at the specified key.
// Parameters:
//
//	ctx - context for the operation
//	key - Redis key for the list
//	value - value to add to the list
//
// Returns:
//
//	error - error if the operation fails, nil otherwise
func AddNotificationToList(ctx context.Context, key string, value string) error {
	return Client.LPush(ctx, key, value).Err()
}

// GetSetMembers retrieves all members of a Redis set at the specified key.
// Parameters:
//
//	ctx - context for the operation
//	key - Redis key for the set
//
// Returns:
//
//	[]string - slice of set members
//	error - error if the operation fails, nil otherwise
func GetSetMembers(ctx context.Context, key string) ([]string, error) {
	return Client.SMembers(ctx, key).Result()
}

// GetListRange retrieves all elements of a Redis list at the specified key.
// Parameters:
//
//	ctx - context for the operation
//	key - Redis key for the list
//
// Returns:
//
//	[]string - slice of list elements
//	error - error if the operation fails, nil otherwise
func GetListRange(ctx context.Context, key string) ([]string, error) {
	return Client.LRange(ctx, key, 0, -1).Result()
}

// DeleteKey deletes the specified key from Redis.
// Parameters:
//
//	ctx - context for the operation
//	key - Redis key to delete
//
// Returns:
//
//	error - error if the operation fails, nil otherwise
func DeleteKey(ctx context.Context, key string) error {
	return Client.Del(ctx, key).Err()
}

// SetKeyExpiry sets an expiration time for the specified Redis key.
// Parameters:
//
//	ctx - context for the operation
//	key - Redis key to set expiration for
//	expiration - duration until the key expires
//
// Returns:
//
//	error - error if the operation fails, nil otherwise
func SetKeyExpiry(ctx context.Context, key string, expiration time.Duration) error {
	return Client.Expire(ctx, key, expiration).Err()
}
