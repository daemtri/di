package httpservice

import (
	"context"
	"net/http"

	"github.com/daemtri/di/example/di_example/clients"
)

type HttpServiceOptions struct {
	Addr string `flag:"addr" default:":8080" usage:"http service listen address"`

	Redis *clients.RedisClient `inject:"must"`
}

func (hso *HttpServiceOptions) Build(ctx context.Context) (*HttpService, error) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := hso.Redis.Get("ping")
		if err != nil {
			w.Write([]byte("error"))
			return
		}
		w.Write([]byte(data))
	})
	return &HttpService{
		opt: hso,
		server: http.Server{
			Addr: hso.Addr,
		},
		ctx: ctx,
	}, nil
}

type HttpService struct {
	opt *HttpServiceOptions

	ctx    context.Context
	server http.Server
}

func (h *HttpService) Run() error {
	go func() {
		<-h.ctx.Done()
		h.server.Shutdown(h.ctx)
	}()
	return h.server.ListenAndServe()
}
