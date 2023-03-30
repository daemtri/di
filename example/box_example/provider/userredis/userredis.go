package userredis

import (
	"github.com/daemtri/di/box"
	"github.com/daemtri/di/example/box_example/client"
)

var (
	Name string = "user"
)

func init() {
	box.Provide[*client.RedisClient](client.NewRedisClient, box.WithName(Name), box.WithFlags("redis-user"))
}
