package users

import "github.com/vivek6201/biolynq/internal/config"

type UserHandler struct {
	service *UserService
	cfg     *config.ConfigVar
}

func NewUserHandler(service *UserService, cfg *config.ConfigVar) *UserHandler {
	return &UserHandler{
		service: service,
		cfg:     cfg,
	}
}
