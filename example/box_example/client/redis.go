package client

import (
	"encoding/json"

	"github.com/daemtri/di/example/box_example/contract"
	"golang.org/x/exp/slog"
)

type RedisOptions struct {
	Host string `flag:"host" default:"127.0.0.1" usage:"redis服务ip地址"`
	Port int    `flag:"port" default:"6379" usage:"redis服务端口"`
	DB   int    `flag:"db" default:"0" usage:"redis db"`
}

type RedisClient struct {
	opts *RedisOptions
}

func (rc *RedisClient) Get(key string) (string, error) {
	data, err := json.Marshal(contract.UserProfile{
		UserID:   key,
		Email:    "xxx@xxx.com",
		Nickname: "xxx",
		Avatar:   "https://xxx.jpeg",
	})
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func NewRedisClient(opts *RedisOptions, logger *slog.Logger) (*RedisClient, error) {
	logger.Info("new redis client", "options", opts)
	return &RedisClient{opts: opts}, nil
}
