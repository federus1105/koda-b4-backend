package libs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// --- GET FROM CACHE ---
func GetFromCache[T any](ctx context.Context, rd *redis.Client, key string) (*T, error) {
	cmd := rd.Get(ctx, key)
	if cmd.Err() != nil {
		if cmd.Err() == redis.Nil {
			return nil, nil
		}
		return nil, cmd.Err()
	}

	var result T
	b, err := cmd.Bytes()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// --- SET TO CACHE IF CACHE NULL ---
func SetToCache[T any](ctx context.Context, rd *redis.Client, key string, value T, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return rd.Set(ctx, key, b, ttl).Err()
}

// --- INVALIDATE CACHE REUSABLE ---
func InvalidateCacheByPattern(ctx context.Context, rd *redis.Client, keyCache string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := rd.Scan(ctx, cursor, keyCache, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := rd.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
