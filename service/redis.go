package service

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/hbahadorzadeh/key-master/util"
	log "github.com/sirupsen/logrus"
)

func NewRedisClient(configs *util.Configs, logger *log.Logger) *redis.Client {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     configs.Redis.Address,
		Password: configs.Redis.Password, // no password set
		DB:       configs.Redis.Index,    // use default DB
	})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		logger.Error(err)
	}
	return rdb
}
