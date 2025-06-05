package config

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
)

var RedisClient *redis.Client
var RedisCtx = context.Background()

func ConnectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // hoặc từ REDIS_URL
		Password: "",               // nếu có mật khẩu thì thêm vào
		DB:       0,
	})

	_, err := RedisClient.Ping(RedisCtx).Result()
	if err != nil {
		log.Fatal("Không thể kết nối Redis:", err)
	}
}
