package main

import (
	"github.com/daemtri/di"
	"github.com/daemtri/di/example/di_example/httpservice"
)

func main() {
	reg := di.NewRegistry()
	di.Provide[*httpservice.HttpService](reg, &httpservice.HttpServiceOptions{})
}
