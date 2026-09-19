package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/avanthika/efootball-backend/internal/models"
	"github.com/avanthika/efootball-backend/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid phone or password")
var ErrPhoneAlreadyRegistered = repository.ErrPhoneAlreadyRegistered

type AuthService interface {
	Login(ctx context.Context, phone, password string) (token string, user *models.User, err error)
	Signup(ctx context.Context, name, phone, password string) (token string, user *models.User, err error)
}

type authService struct {
	users     repository.UserRepository
	jwtSecret string
}

func NewAuthService(users repository.UserRepository, jwtSecret string) AuthService {
	return &authService{users: users, jwtSecret: jwtSecret}
}

func (s *authService) Login(ctx context.Context, phone, password string) (string, *models.User, error) {
	user, err := s.users.FindByPhone(ctx, phone)
	if errors.Is(err, repository.ErrUserNotFound) {
		return "", nil, ErrInvalidCredentials
	}
	if err != nil {
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	signed, err := s.issueToken(user)
	if err != nil {
		return "", nil, err
	}
	return signed, user, nil
}

func (s *authService) Signup(ctx context.Context, name, phone, password string) (string, *models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}

	// Public signup can only ever create a subadmin — role is never
	// accepted from the request, it is fixed here.
	user, err := s.users.Create(ctx, name, phone, string(hash), models.RoleSubadmin)
	if err != nil {
		return "", nil, err
	}

	signed, err := s.issueToken(user)
	if err != nil {
		return "", nil, err
	}
	return signed, user, nil
}

func (s *authService) issueToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"userId": user.ID,
		"role":   user.Role,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
