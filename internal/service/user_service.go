package service

import (
	"context"

	"github.com/ElissaDesign/vpn-dev/internal/model"
	"github.com/ElissaDesign/vpn-dev/internal/repository"
	"github.com/google/uuid"
)

// UserService handles business logic for users
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetUsers retrieves all users
func (s *UserService) GetUsers(ctx context.Context) ([]model.User, error) {
	return s.repo.GetAll(ctx)
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, user *model.User) error {
	return s.repo.Create(ctx, user)
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.repo.FindByID(ctx, id)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.repo.FindByEmail(ctx, email)
}

// GetUserByIP retrieves a user by IP address
func (s *UserService) GetUserByIP(ctx context.Context, ip string) (*model.User, error) {
	return s.repo.FindByIP(ctx, ip)
}

// GetUserByPublicKey retrieves a user by WireGuard public key
func (s *UserService) GetUserByPublicKey(ctx context.Context, publicKey string) (*model.User, error) {
	return s.repo.FindByPublicKey(ctx, publicKey)
}

// UpdateUserStatus updates a user's connection status
func (s *UserService) UpdateUserStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

// UpdateLastConnected updates the last connection time
func (s *UserService) UpdateLastConnected(ctx context.Context, id uuid.UUID) error {
	return s.repo.UpdateLastConnected(ctx, id)
}

// DeleteUser deletes a user (used for transaction rollback)
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
