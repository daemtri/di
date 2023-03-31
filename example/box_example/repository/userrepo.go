package repository

import (
	"encoding/json"
	"fmt"

	"github.com/daemtri/di/example/box_example/client"
	"github.com/daemtri/di/example/box_example/contract"
	"golang.org/x/exp/slog"
)

type UserRedisRepository struct {
	c      *client.RedisClient
	logger *slog.Logger
}

type Kaka struct{}

func NewUserRedisRepository(c *client.RedisClient, ka *Kaka, logger *slog.Logger) (*UserRedisRepository, error) {
	return &UserRedisRepository{c: c, logger: logger}, nil
}

func (u *UserRedisRepository) GetUserProfile(userid string) (*contract.UserProfile, error) {
	u.logger.Info("get user profile", "userid", userid)
	rt, err := u.c.Get(fmt.Sprintf("user:%s:profile", userid))
	if err != nil {
		return nil, err
	}
	var profile contract.UserProfile
	if err := json.Unmarshal([]byte(rt), &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}
