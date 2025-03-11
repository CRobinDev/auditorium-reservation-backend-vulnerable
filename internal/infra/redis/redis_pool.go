package redis

import (
	"context"
	"fmt"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/log"
	"github.com/redis/go-redis/v9"
	"sync"
)

var (
	client *redis.Client
	once   sync.Once
)

func NewRedisPool(host, port, pass string, db int) *redis.Client {
	once.Do(func() {
		cl := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: pass,
			DB:       db,
		})

		ping, err := cl.Ping(context.Background()).Result()
		if err != nil {
			log.Fatal(map[string]interface{}{
				"error": err.Error(),
			}, "[REDIS][NewRedisPool] failed to connect to redis")
		}

		log.Info(map[string]interface{}{
			"ping": ping,
		}, "[REDIS][NewRedisPool] connected to redis")

		client = cl
	})

	return client
}
