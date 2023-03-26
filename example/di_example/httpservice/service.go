package httpservice

import (
	"context"
	"net/http"
)

type HttpServiceOptions struct {
	Addr string `flag:"addr" default:":8080" usage:"http service listen address"`
}

func (hso *HttpServiceOptions) Build(ctx context.Context) (*HttpService, error) {
	return &HttpService{
		addr: hso.Addr,
		server: http.Server{
			Addr: hso.Addr,
		},
	}, nil
}

type HttpService struct {
	addr   string
	server http.Server
}

func (h *HttpService) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		h.server.Shutdown(ctx)
	}()
	return h.server.ListenAndServe()
}
