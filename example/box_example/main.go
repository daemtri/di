package main

import (
	"github.com/daemtri/di/box"
	"github.com/daemtri/di/box/config/apolloconfig"
	"github.com/daemtri/di/box/config/yamlconfig"
	"github.com/daemtri/di/example/box_example/bootstrap"
	"github.com/daemtri/di/example/box_example/client"
	"github.com/daemtri/di/example/box_example/contract"
	"github.com/daemtri/di/example/box_example/provider/userredis"
	"github.com/daemtri/di/example/box_example/repository"
	"github.com/daemtri/di/example/box_example/server"
	"github.com/daemtri/di/example/box_example/service"
	"golang.org/x/exp/slog"
)

func main() {
	// register redis client
	box.Provide[contract.UserRepository](
		repository.NewUserRedisRepository,
		box.WithSelect[*client.RedisClient](userredis.Name),
	)
	// register service
	box.Provide[contract.Service](service.NewUserService, box.WithName("user"))
	box.Provide[contract.Service](service.NewConsumerService, box.WithName("consumer"))
	// register server
	box.Provide[contract.Server](server.NewHttpServer, box.WithName("http"), box.WithFlags("http"))
	// register logger
	box.Provide[*slog.Logger](slog.Default())
	// register app
	box.Provide[*bootstrap.App](bootstrap.NewApp)

	// build and run
	if err := box.Bootstrap[*bootstrap.App](
		// The configuration has priority, the higher the priority of the closer
		yamlconfig.Init(),
		apolloconfig.Init(),
	); err != nil {
		panic(err)
	}
}
