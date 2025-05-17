package databases

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func CreateClient(dbNum int) *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = os.Getenv("DB_PORT")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASS"),
		DB:       dbNum,
	})
	return rdb
}
