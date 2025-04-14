package auth

import (
	"context"
	"errors"
	"fmt"

	"pvz_service/internal/storage/postgres"
)

// mockgen -source=internal/app/auth/repository.go -destination=internal/app/auth/mocks/repository_mock.go -package=mocks

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	CheckUser(ctx context.Context, registerRequest *RegisterRequest) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type Repo struct {
	Storage *postgres.Storage
}

func NewUserRepository(storage *postgres.Storage) *Repo {
	return &Repo{
		Storage: storage,
	}
}

func (u *Repo) CreateUser(ctx context.Context, user *User) error {
	query := `
			INSERT INTO users (id, email, password, role) 
			VALUES ($1, $2, $3, $4)
			`

	_, err := u.Storage.DB.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Role,
	)
	if err != nil {
		return fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	return nil
}

func (u *Repo) CheckUser(ctx context.Context, registerRequest *RegisterRequest) error {
	var countUser int

	query := "SELECT COUNT(*) FROM users WHERE email = $1"

	err := u.Storage.DB.QueryRowContext(
		ctx,
		query,
		registerRequest.Email,
	).Scan(&countUser)
	if err != nil {
		return fmt.Errorf("произошла ошибка проверки пользователя: %w", err)
	}

	if countUser != 0 {
		return errors.New("пользователь уже существует")
	}

	return nil
}

func (u *Repo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	query := `
			SELECT id, email, password, role, created_at
			FROM users WHERE email = $1
			`

	rows := u.Storage.DB.QueryRowContext(
		ctx,
		query,
		email,
	)

	err := rows.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("произошла ошибка получения пользователя: %w", err)
	}

	return &user, nil
}
