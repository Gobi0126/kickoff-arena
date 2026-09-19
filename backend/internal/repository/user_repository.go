package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avanthika/efootball-backend/internal/models"
)

var ErrUserNotFound = errors.New("user not found")
var ErrPhoneAlreadyRegistered = errors.New("phone already registered")
var ErrUserHasDependents = errors.New("cannot delete: this account still owns other records")

type UserRepository interface {
	FindByPhone(ctx context.Context, phone string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	Create(ctx context.Context, name, phone, passwordHash string, role models.Role) (*models.User, error)
	ListByRole(ctx context.Context, role models.Role) ([]models.User, error)
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, phone, password_hash, role, created_at
		 FROM users WHERE phone = $1`,
		phone,
	)

	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Phone, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, phone, password_hash, role, created_at
		 FROM users WHERE id = $1`,
		id,
	)

	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Phone, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) ListByRole(ctx context.Context, role models.Role) ([]models.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, phone, password_hash, role, created_at
		 FROM users WHERE role = $1 ORDER BY created_at DESC`,
		role,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Phone, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrUserHasDependents
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *userRepository) Create(ctx context.Context, name, phone, passwordHash string, role models.Role) (*models.User, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO users (name, phone, password_hash, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, phone, password_hash, role, created_at`,
		name, phone, passwordHash, role,
	)

	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Phone, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrPhoneAlreadyRegistered
		}
		return nil, err
	}
	return &u, nil
}
