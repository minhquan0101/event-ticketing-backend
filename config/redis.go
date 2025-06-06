package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var RedisCtx = context.Background()

func ConnectRedis() {
	redisURL := os.Getenv("REDIS_URL")

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("Không thể parse REDIS_URL:", err)
	}

	RedisClient = redis.NewClient(opt)

	_, err = RedisClient.Ping(RedisCtx).Result()
	if err != nil {
		log.Fatal("Không thể kết nối Redis:", err)
	}

	log.Println("✅ Đã kết nối Redis thành công")
}
