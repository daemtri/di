package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/daemtri/di"
	"github.com/daemtri/di/box/flagx"
	"github.com/daemtri/di/example/di_example/clients"
	"github.com/daemtri/di/example/di_example/httpservice"
)

func main() {
	nfs := flagx.NamedFlagSets{}
	di.Provide[*httpservice.HttpService](&httpservice.HttpServiceOptions{}, di.WithFlagset(nfs.FlagSet("httpservice")))
	di.Provide[*clients.RedisClient](&clients.RedisOptions{}, di.WithFlagset(nfs.FlagSet("redis")))
	di.Provide[*clients.MysqlClient](&clients.MysqlOptions{}, di.WithFlagset(nfs.FlagSet("mysql")))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	server, err := di.Build[*httpservice.HttpService](ctx)
	if err != nil {
		panic(err)
	}
	if err := server.Run(); err != nil {
		panic(err)
	}
}
