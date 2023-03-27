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
	di.Provide[*clients.RedisClient](
		&clients.RedisOptions{},
		di.WithFlagset(nfs.FlagSet("redis")),
	)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	server, err := di.Build[*httpservice.HttpService](ctx,
		&httpservice.HttpServiceOptions{},
		di.WithFlagset(nfs.FlagSet("httpservice")))
	if err != nil {
		panic(err)
	}
	if err := server.Run(); err != nil {
		panic(err)
	}
}
