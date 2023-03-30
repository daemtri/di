package service

import (
	"encoding/json"
	"net/http"

	"github.com/daemtri/di/example/box_example/contract"
	"golang.org/x/exp/slog"
)

type UserService struct {
	repo   contract.UserRepository
	logger *slog.Logger
}

func NewUserService(repo contract.UserRepository, logger *slog.Logger) (*UserService, error) {
	return &UserService{repo: repo, logger: logger}, nil
}

func (u *UserService) AddRoute(h *http.ServeMux) {
	u.logger.Info("UserService add Route")
	h.HandleFunc("/user/profile", u.UserProfile)
}

func (u *UserService) UserProfile(w http.ResponseWriter, r *http.Request) {
	slog.Info("user service handle user profile")
	profile, err := u.repo.GetUserProfile(r.URL.Query().Get("userid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(profile)
}
