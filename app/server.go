package main

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/redis"
	// Uncomment this block to pass the first stage
)

func main() {
	fmt.Println("Logs from your program will appear here!")
	redisServer := redis.NewRedisServer()
	redisServer.StartRedis()
}