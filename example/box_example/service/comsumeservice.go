package service

import (
	"net/http"

	"github.com/daemtri/di/example/box_example/contract"
	"golang.org/x/exp/slog"
)

type ConsumerService struct {
	logger *slog.Logger
}

func NewConsumerService(repo contract.UserRepository, logger *slog.Logger) (*ConsumerService, error) {
	return &ConsumerService{logger: logger}, nil
}

func (c *ConsumerService) AddRoute(mux *http.ServeMux) {
	c.logger.Info("ConsumerService add Route")
}
