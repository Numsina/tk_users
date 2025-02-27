package initialize

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {
	addr := fmt.Sprintf("%s:%d", Conf.RedisInfo.Host, Conf.RedisInfo.Port)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: Conf.RedisInfo.PassWord,
	})
	return redisClient
}
