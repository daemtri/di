package server

import (
	"net/http"

	"github.com/daemtri/di/example/box_example/contract"
	"golang.org/x/exp/slog"
)

// HttpServerOptions http服务配置
type HttpServerOptions struct {
	Addr string `flag:"addr" default:":10890" usage:"http服务监听地址"`
}

// NewHttpServer 创建http服务
func NewHttpServer(opt *HttpServerOptions, services []contract.Service, logger *slog.Logger) (*http.Server, error) {
	mux := http.NewServeMux()
	for i := range services {
		services[i].AddRoute(mux)
	}
	s := &http.Server{
		Addr:     opt.Addr,
		Handler:  mux,
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
	slog.Info("create http server succeed", "listen", opt.Addr)
	return s, nil
}
