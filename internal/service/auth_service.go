package service

import (
	"context"

	"github.com/ElissaDesign/vpn-dev/internal/model"
	"github.com/ElissaDesign/vpn-dev/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUsers(ctx context.Context) ([]model.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *UserService) CreateUser(ctx context.Context, user *model.User) error {
	_, err := s.repo.Create(ctx, user)
	return err
}

func (s *UserService) GetUserByPublicKey(ctx context.Context, publicKey string) (*model.User, error) {
	return s.repo.FindByPublicKey(ctx, publicKey)
}
