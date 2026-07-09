package main

import (
	"fmt"
	"log"

	redisdb "tutproj/internal/repository/redis"
)

func main() {
	rdb := redisdb.CreateConnection()
	defer rdb.Close() // FIX 1: Close here, after we're done using the client

	redisdb.AddToken("my-token", rdb)

	exists, err := redisdb.CheckToken("my-token", rdb)
	if err != nil {
		log.Fatalf("Error checking token: %v", err)
	}
	fmt.Printf("Token exists: %v\n", exists)
}
