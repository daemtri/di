package clients

import (
	"context"
	"fmt"
)

type MysqlOptions struct {
	Addr string `flag:"addr" default:"127.0.0.1:3306" usage:"mysql server address"`
}

func (mo *MysqlOptions) Build(ctx context.Context) (*MysqlClient, error) {
	fmt.Println("build mysql client", mo.Addr)
	// 这里可以用mysql驱动的client
	return &MysqlClient{client: nil, Addr: mo.Addr}, nil
}

type MysqlClient struct {
	Addr   string
	client any // 这里可以是mysql驱动的client
}
