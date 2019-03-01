package db

import (
	"go_task/common"
	"github.com/garyburd/redigo/redis"
	"time"
)

var RedisPool *redis.Pool

func InitRedis() error {
	var err error
	RedisPool, err = newPool()

	return err
}

func newPool() (pool *redis.Pool, err error) {

	pool = &redis.Pool{
		MaxActive:   200,
		MaxIdle:     100,
		IdleTimeout: 120 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", common.Redisaddr)
			if err != nil {
				return nil, err
			}

			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return pool, nil
}