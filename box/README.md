# Box - 容器库

box 库提供了一个全局的`依赖注入容器`，在初始化阶段注册你的所有依赖，在Build阶段，根据依赖关系，递归构建依赖

| 注意：依赖注入容器只支持单例模式，如果需要多例，可注册一个类型的Manager，实现多例

[toc]

## 快速开始

```go
package main

import (
	"github.com/daemtri/di/box"
	"github.com/daemtri/di/logx"
)

log = logx.GetLogger("main")

func main() {
	// 使用 box.Provide 注入类型构造器
	// 参数需要实现 box.Builder[T] 接口
	box.Provide[*Mysql.Client](&mysql.Builder{}, box.WithFlagPrefix("mysql"))
	// app.ServerBuilder可通过 di.Must[*mysql.Client](ctx) 获取mysql依赖
	box.Provide[*app.Server](&app.ServerBuilder{})

	// 构建依赖，Build只能执行一次
	srv, err := box.Build[*app.Server](context.TODO())
	if err != nil {
		log.Panic("build error", "error", err)
	}
	srv.Run()
}
```

## Provide 语法糖

### 使用 ProvideFunc 注入构造函数

```go
package main

import (
	"flag"
	"github.com/daemtri/di/box"
)

type Server struct {
	mysql *mysql.Client
}

func (s *Server) Run() {}

// NewServer Option由Build函数使用发射创建，不需要Provide
// ServerRunOption 需要实现接口
func NewServer(ctx box.Context) (*Server, error) {
	// 使用box.Must获取依赖
	mysqlClient := box.Must[*mysql.Client](ctx)
	return &Server{mysql: mysqlClient}, nil
}

func main() {
	// ProvideFunc 注入构造函数
	// 参数签名为: func(box.Context) (T, error)
	box.ProvideFunc[*Server](NewServer)
}
```

### 使用 ProvideOptionFunc 注入带Option的构造函数

```go
package main

import (
	"flag"
	"github.com/daemtri/di/box"
)

type ServerRunOption struct {
	Addr string
}

func (s *ServerRunOption) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&addr, "addr", ":80", "服务监听地址")
}

func (s *ServerRunOption) ValidateFlags() error {
	// 验证 addr 是否合法
	return nil
}

type Server struct {
	mysql *mysql.Client
}
 
func (s *Server) Run() {}

// NewServer Option由Build函数使用发射创建，不需要Provide
// ServerRunOption 需要实现接口
func NewServer(ctx box.Context, opt *ServerRunOption) (*Server, error) {
	// 使用box.Must获取依赖
	mysqlClient := box.Must[*mysql.Client](ctx)
	return &Server{mysql: mysqlClient}, nil
}

func main() {
	// ProvideOptionFunc 注入构造函数
	// 参数签名为: func(box.Context) (T, error)
	box.ProvideOptionFunc[*Server](NewServer)
}
```

### 使用 ProvideInject 注入动态参数构造函数 （推荐）

```go
package main

import (
	"github.com/daemtri/di/box"
)

type ServerRunOption struct {
	Addr string `flag:"addr" default:":80" usage:"服务监听地址" validate:"host_port"`
}

type Server struct {
	ctx   context.Context
	mysql *mysql.Client
}

func (s *Server) Run() {}

// NewServer context.Context,ServerRunOption  不需要Provide，由Build函数传入
// *mysql.Client 需要Provide后，才能获取到
func NewServer(ctx context.Context, mysqlClient *mysql.Client, opt *ServerRunOption) (*Server, error) {
	return &Server{ctx ctx, mysql: mysqlClient}, nil
}

func main() {
	// ProvideInject 注入自动构造函数
	// mysql.NewClient仍然是一个函数, 其签名为: func(arg1 arg1Type ...) (T, error)
	box.ProvideInject[*Server](NewServer)
}
```

### 使用 ProvideInstance 直接注入对象

```go
package main

import (
	"github.com/daemtri/di/box"
)

type Server struct {
	Addr string `flag:"addr" default:":80" usage:"服务监听地址" validate:"host_port"`

	Ctx   context.Context `inject:"exists"` //  存在时注入
	Mysql *mysql.Client   `inject:"must"`   // 必须注入
}

func (s *Server) Run() {}

func main() {
	box.ProvideInstance[*Server](&Server{})
}
```

## 参数说明

```
-config 指定参数路径
-print-config 打印当前使用的参数配置
```

## FAQ

### 使用box.WithFlagPrecix("server","run")设置类型的参数前缀避免多个类型参数冲突

### 使用box.WithName("name")支持Provide多个相同类型

### ProvideInject和ProvideInstance如何传递命名参数?

### 使用box.WithOverride()覆盖已经Provided的类型