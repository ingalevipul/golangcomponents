package redisdb

import (
	"context"
	"errors"
	"fmt"

	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func CreateConnection() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0, // Default database ID
	})

	err := rdb.Ping(ctx).Err()
	if err != nil {
		fmt.Printf("Could not connect to Redis: %v", err)
	}
	fmt.Println("Successfully connected to Redis!")

	return rdb
}

func AddToken(token string, rdb *redis.Client) {
	err := rdb.Set(ctx, token, true, 7*24*time.Hour).Err()
	if err != nil {
		fmt.Printf("Failed to SET key: %v", err)
	}
}

func CheckToken(token string, rdb *redis.Client) (bool, error) {
	_, err := rdb.Get(ctx, token).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		// A genuine Redis error
		return false, fmt.Errorf("failed to GET key: %w", err)
	}
	return true, nil
}
