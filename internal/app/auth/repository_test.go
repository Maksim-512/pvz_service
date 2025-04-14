package auth_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pvz_service/internal/app/auth"
	"pvz_service/internal/storage/postgres"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := auth.NewUserRepository(&postgres.Storage{DB: db})

	userID := uuid.New()
	user := &auth.User{
		ID:           userID,
		Email:        "test@yandex.ru",
		PasswordHash: "password",
		Role:         auth.RoleEmployee,
	}

	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO users (id, email, password, role) 
			VALUES ($1, $2, $3, $4)`),
	).
		WithArgs(user.ID, user.Email, user.PasswordHash, user.Role).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateUser(context.Background(), user)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestCheckUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := auth.NewUserRepository(&postgres.Storage{DB: db})

	registerRequest := &auth.RegisterRequest{
		Email:    "test@yandex.ru",
		Password: "password",
		Role:     auth.RoleEmployee,
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT COUNT(*) 
			FROM users 
			WHERE email = $1`),
	).
		WithArgs(registerRequest.Email).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	err = repo.CheckUser(context.Background(), registerRequest)
	assert.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT COUNT(*) 
			FROM users 
			WHERE email = $1`),
	).
		WithArgs(registerRequest.Email).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	err = repo.CheckUser(context.Background(), registerRequest)
	assert.Error(t, err)
	assert.Equal(t, "пользователь уже существует", err.Error())

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := auth.NewUserRepository(&postgres.Storage{DB: db})

	userEmail := "test@yandex.ru"
	expectedUser := &auth.User{
		ID:           uuid.New(),
		Email:        userEmail,
		PasswordHash: "password",
		Role:         auth.RoleEmployee,
		CreatedAt:    time.Now(),
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, email, password, role, created_at 
			FROM users 
			WHERE email = $1`),
	).
		WithArgs(userEmail).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at"}).
			AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PasswordHash, expectedUser.Role, expectedUser.CreatedAt))

	user, err := repo.GetUserByEmail(context.Background(), userEmail)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.Email, user.Email)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestCreateUser_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := auth.NewUserRepository(&postgres.Storage{DB: db})

	userID := uuid.New()
	user := &auth.User{
		ID:           userID,
		Email:        "test@yandex.ru",
		PasswordHash: "password",
		Role:         auth.RoleEmployee,
	}

	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO users (id, email, password, role) 
			VALUES ($1, $2, $3, $4)`),
	).
		WithArgs(user.ID, user.Email, user.PasswordHash, user.Role).
		WillReturnError(fmt.Errorf("произошла ошибка при выполнении запроса"))

	err = repo.CreateUser(context.Background(), user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка создания пользователя")
}

func TestCheckUser_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := auth.NewUserRepository(&postgres.Storage{DB: db})

	registerRequest := &auth.RegisterRequest{
		Email:    "test@yandex.ru",
		Password: "password",
		Role:     auth.RoleEmployee,
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT COUNT(*) 
			FROM users 
			WHERE email = $.+`),
	).
		WithArgs(registerRequest.Email).
		WillReturnError(fmt.Errorf("произошла ошибка при выполнении запроса"))

	err = repo.CheckUser(context.Background(), registerRequest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "произошла ошибка проверки пользователя")
}

func TestGetUserByEmail_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := auth.NewUserRepository(&postgres.Storage{DB: db})

	email := "test@yandex.ru"

	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, email, password, role, created_at 
			FROM users 
			WHERE email = $1`),
	).
		WithArgs(email).
		WillReturnError(fmt.Errorf("произошла ошибка при выполнении запроса"))

	_, err = repo.GetUserByEmail(context.Background(), email)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "произошла ошибка получения пользователя")
}
