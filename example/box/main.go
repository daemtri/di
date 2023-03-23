package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/daemtri/di/box"
	"golang.org/x/exp/slog"
)

type UserRepo interface {
	GetEmail(id string) string
}

type UserRedisRepo struct {
	// client
}

func (urp *UserRedisRepo) GetEmail(id string) string {
	return "myname@xxx.com"
}

type UserRedisRepoOptions struct {
	Host     string `flag:"host" default:"127.0.0.1" usage:"ip地址"`
	Port     int    `flag:"port" default:"6379" usage:"端口"`
	Username string `flag:"username" default:"myname" usage:"用户名"`
	Password string `flag:"password" default:"mypassword" usage:"密码"`
}

func NewUserRedisRepo(opt *UserRedisRepoOptions) (*UserRedisRepo, error) {
	// 使用opt初始化client
	slog.Info("init redis client", "options", opt)
	return &UserRedisRepo{}, nil
}

type HttpServer struct {
	server   *http.Server
	userRepo UserRepo
}

type HttpServerRunOptions struct {
	Addr string `flag:"addr" default:"0.0.0.0:8088" usage:"http服务监听地址"`
}

func NewHttpServer(opt *HttpServerRunOptions, repo UserRepo) (*HttpServer, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/email", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		w.Write([]byte(repo.GetEmail(id)))
	})

	return &HttpServer{
		userRepo: repo,
		server: &http.Server{
			Addr:    opt.Addr,
			Handler: mux,
		},
	}, nil
}

func (hs HttpServer) Run(ctx context.Context) error {
	slog.Info("服务已启动 ")
	go func() {
		<-ctx.Done()
		sCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := hs.server.Shutdown(sCtx); err != nil {
			slog.Warn("shutdown http server", "error", err)
		}
	}()
	return hs.server.ListenAndServe()
}

func main() {
	box.Provide[UserRepo](NewUserRedisRepo, box.WithFlagPrefix("repo-user-redis"))
	box.Provide[*HttpServer](NewHttpServer, box.WithFlagPrefix("server-http"))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	server, err := box.Build[*HttpServer](ctx, box.UseInit(func(ctx box.Context) error {
		slog.Info("这里在build之前执行,可以做一些初始的工作，比如日志库，远程配置等等")
		return nil
	}))
	if err != nil {
		slog.Error("build失败", "error", err)
	}
	if err := server.Run(ctx); err != nil {
		slog.Error("服务退出", "error", err)
	}
}
