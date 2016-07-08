// arkadiko
// https://github.com/topfreegames/arkadiko
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package redisclient

import (
	"fmt"
	"sync"

	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

var once sync.Once
var client *RedisClient

type RedisClient struct {
	Logger zap.Logger
	Pool   *redis.Pool
}

func GetRedisClient() *RedisClient {
	once.Do(func() {
		client = &RedisClient{
			Logger: zap.NewJSON(zap.WarnLevel),
		}
		redisAddress := fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port"))
		client.Pool = redis.NewPool(func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisAddress)
			if err != nil {
				if err != nil {
					client.Logger.Error(err.Error())
				}
			}
			return c, err
		}, viper.GetInt("redis.maxPoolSize"))
	})
	return client
}
