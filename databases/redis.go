package databases

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func CreateClient(dbNum int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		DB:       dbNum,
	})
	return rdb
}
