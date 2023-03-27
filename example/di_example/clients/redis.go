package clients

import "context"

type RedisOptions struct {
	Addr string `flag:"addr" default:"127.0.0.1" usage:"redis server address"`
	Port int    `flag:"port" default:"6379" usage:"redis server port"`
}

func (ro *RedisOptions) Build(ctx context.Context) (*RedisClient, error) {
	// 这里可以用redis驱动的client
	return &RedisClient{client: nil}, nil
}

type RedisClient struct {
	client any // 这里可以是redis驱动的client
}

func (rc *RedisClient) Get(key string) (string, error) {
	return "test redis client get", nil
}
