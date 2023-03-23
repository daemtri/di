# DI 容器

## 使用说明


```go
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/daemtri/di/pkg/di"
	"github.com/daemtri/di/pkg/flagx"
	"log"
)

// -----------------------------------------GameService------------------------------------------------------

type GameService interface {
	Run()
}

type gameServiceImpl struct {
	redisClient *RedisClient
}

func (g *gameServiceImpl) Run() {
	fmt.Println("hello")
	fmt.Println(g.redisClient.Key("gameServiceImpl"))
}

type GameServiceImplBuilder struct {
	di.NoneFlags
}

func (g *GameServiceImplBuilder) Build(ctx di.Context) (GameService, error) {
	return &gameServiceImpl{
		redisClient: di.Must[*RedisClient](ctx.Select("main")),
	}, nil
}

// -----------------------------------------Redis------------------------------------------------------

type RedisClient struct {
	addr string
}

func (rc *RedisClient) Key(k string) string {
	return fmt.Sprintf("addr:%s -> %s", rc.addr, k)
}

type RedisClientBuilder struct {
	addr     string
	port     int
	user     string
	password string
}

func (r *RedisClientBuilder) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&r.addr, "addr", "127.0.0.1:6479", "redis地址")
}

func (r *RedisClientBuilder) ValidateFlags() error {
	// TODO: 验证addr等参数合法性
	return nil
}

func (r *RedisClientBuilder) Build(ctx di.Context) (*RedisClient, error) {
	return &RedisClient{addr: r.addr}, nil
}

// -----------------------------------------main------------------------------------------------------

func main() {
	var fs flagx.NamedFlagSets

	c := di.New()
	di.Provide[GameService](c, &GameServiceImplBuilder{})
	
	// Provide 命名对象（当前命名对象不能直接使用di.Build构建）
	di.Provide[*RedisClient](c, di.Named[*RedisClient]("main", &RedisClientBuilder{})).AddFlags(fs.FlagSet("redis-main"))
	di.Provide[*RedisClient](c, di.Named[*RedisClient]("game", &RedisClientBuilder{})).AddFlags(fs.FlagSet("redis-game"))
	
	// 直接Provide 实例,不需要构造器
	di.Provide[bool](c,di.Instance(true))
	// 使用di.Func 简化构造器
	di.Provide[int](c,di.Func(func(c di.Context) (int, error) {
        return 100,nil
	}))

	// 参数必须在build之前解析
	fs.Parse()

	gs, err := di.Build[GameService](c, context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	gs.Run()
}

```
