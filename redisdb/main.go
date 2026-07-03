package redisdb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Background context for Redis commands
var ctx = context.Background()

func CreateConnection() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis host and port
		Password: "",               // No password by default
		DB:       0,                // Default database ID
	})

	err := rdb.Ping(ctx).Err()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	fmt.Println("Successfully connected to Redis!")

	// FIX 1: Do NOT close here. The caller owns the client lifecycle.
	return rdb
}

func AddToken(token string, rdb *redis.Client) {
	err := rdb.Set(ctx, token, true, 7*24*time.Hour).Err()
	if err != nil {
		log.Fatalf("Failed to SET key: %v", err)
	}
}

// FIX 2: Returns (bool, error) — handles cache miss (redis.Nil) without crashing.
func CheckToken(token string, rdb *redis.Client) (bool, error) {
	_, err := rdb.Get(ctx, token).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Key does not exist — normal case, not a real error
			return false, nil
		}
		// A genuine Redis error
		return false, fmt.Errorf("failed to GET key: %w", err)
	}
	return true, nil
}

func main() {
	rdb := CreateConnection()
	defer rdb.Close() // FIX 1: Close here, after we're done using the client

	AddToken("my-token", rdb)

	exists, err := CheckToken("my-token", rdb)
	if err != nil {
		log.Fatalf("Error checking token: %v", err)
	}
	fmt.Printf("Token exists: %v\n", exists)
}
