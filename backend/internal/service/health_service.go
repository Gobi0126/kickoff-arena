package service

import (
	"context"

	"github.com/avanthika/efootball-backend/internal/repository"
)

type HealthService interface {
	Check(ctx context.Context) (bool, error)
}

type healthService struct {
	repo repository.HealthRepository
}

func NewHealthService(repo repository.HealthRepository) HealthService {
	return &healthService{repo: repo}
}

func (s *healthService) Check(ctx context.Context) (bool, error) {
	if err := s.repo.Ping(ctx); err != nil {
		return false, err
	}
	return true, nil
}
