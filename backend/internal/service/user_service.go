package service

import (
	"context"

	"github.com/avanthika/efootball-backend/internal/models"
	"github.com/avanthika/efootball-backend/internal/repository"
)

type UserService interface {
	ListSubadmins(ctx context.Context) ([]models.User, error)
	GetSubadmin(ctx context.Context, id string) (*models.User, error)
	DeleteSubadmin(ctx context.Context, id string) error
}

type userService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) UserService {
	return &userService{users: users}
}

func (s *userService) ListSubadmins(ctx context.Context) ([]models.User, error) {
	return s.users.ListByRole(ctx, models.RoleSubadmin)
}

func (s *userService) GetSubadmin(ctx context.Context, id string) (*models.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.Role != models.RoleSubadmin {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func (s *userService) DeleteSubadmin(ctx context.Context, id string) error {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user.Role != models.RoleSubadmin {
		return repository.ErrUserNotFound
	}
	return s.users.Delete(ctx, id)
}
