package reception_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pvz_service/internal/app/reception"
	"pvz_service/internal/storage/postgres"
)

func setupTestRepo(t *testing.T) (*reception.Repo, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	storage := &postgres.Storage{DB: db}
	repo := reception.NewReceptionRepository(storage)

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestCreateReception_Success(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	pvzID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO reception (id, pvz_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, pvz_id, datetime, status;
	`)).
		WithArgs(receptionID, pvzID, reception.StatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"id", "pvz_id", "datetime", "status"}).
			AddRow(receptionID, pvzID, now, reception.StatusInProgress))

	rcp := &reception.Reception{
		ID:    receptionID,
		PvzID: pvzID,
	}

	created, err := repo.CreateReception(ctx, rcp)

	require.NoError(t, err)
	assert.Equal(t, receptionID, created.ID)
	assert.Equal(t, pvzID, created.PvzID)
	assert.Equal(t, reception.StatusInProgress, created.Status)
}

func TestCreateReception_Error(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	pvzID := uuid.New()
	receptionID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO reception (id, pvz_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, pvz_id, datetime, status;
	`)).
		WithArgs(receptionID, pvzID, reception.StatusInProgress).
		WillReturnError(errors.New("ошибка добавления"))

	rcp := &reception.Reception{
		ID:    receptionID,
		PvzID: pvzID,
	}

	created, err := repo.CreateReception(ctx, rcp)
	require.Error(t, err)
	assert.Nil(t, created)
}

func TestCheckValidPVZID_Exists(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	pvzID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM pvz
		WHERE id = $1
	`)).
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	ok, err := repo.CheckValidPVZID(context.Background(), pvzID)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCheckValidPVZID_NotExists(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	pvzID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM pvz
		WHERE id = $1
	`)).
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	ok, err := repo.CheckValidPVZID(context.Background(), pvzID)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckValidPVZID_Error(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	pvzID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM pvz
		WHERE id = $1
	`)).
		WithArgs(pvzID).
		WillReturnError(errors.New("неверный запрос"))

	ok, err := repo.CheckValidPVZID(context.Background(), pvzID)
	require.Error(t, err)
	assert.False(t, ok)
}

func TestCheckValidPVZID_ErrNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &postgres.Storage{DB: db}
	repo := reception.NewReceptionRepository(storage)

	pvzID := uuid.New()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM pvz WHERE id = \\$1").
		WithArgs(pvzID).
		WillReturnError(sql.ErrNoRows)

	ok, err := repo.CheckValidPVZID(context.Background(), pvzID)

	require.NoError(t, err)
	require.False(t, ok)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckOpenReception_Found(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	pvzID := uuid.New()
	receptionID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id
		FROM reception
		WHERE pvz_id = $1 and status = $2
	`)).
		WithArgs(pvzID, reception.StatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(receptionID))

	found, id, err := repo.CheckOpenReception(context.Background(), pvzID)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, receptionID, id)
}

func TestCheckOpenReception_NotFound(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	pvzID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id
		FROM reception
		WHERE pvz_id = $1 and status = $2
	`)).
		WithArgs(pvzID, reception.StatusInProgress).
		WillReturnError(sql.ErrNoRows)

	found, id, err := repo.CheckOpenReception(context.Background(), pvzID)
	require.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, uuid.Nil, id)
}

func TestCheckOpenReception_Error(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	pvzID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id
		FROM reception
		WHERE pvz_id = $1 and status = $2
	`)).
		WithArgs(pvzID, reception.StatusInProgress).
		WillReturnError(errors.New("неверный запрос"))

	found, id, err := repo.CheckOpenReception(context.Background(), pvzID)
	require.Error(t, err)
	assert.False(t, found)
	assert.Equal(t, uuid.Nil, id)
}

func TestCloseReception_Success(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	receptionID := uuid.New()
	pvzID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`
			UPDATE reception 
			SET status = $1
			WHERE id = $2
			RETURNING id, pvz_id, datetime, status;
	`)).
		WithArgs(reception.StatusClosed, receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "pvz_id", "datetime", "status"}).
			AddRow(receptionID, pvzID, now, reception.StatusClosed))

	closed, err := repo.CloseReception(ctx, receptionID)
	require.NoError(t, err)
	assert.Equal(t, receptionID, closed.ID)
	assert.Equal(t, reception.StatusClosed, closed.Status)
}

func TestCloseReception_Error(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	receptionID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
			UPDATE reception 
			SET status = $1
			WHERE id = $2
			RETURNING id, pvz_id, datetime, status;
	`)).
		WithArgs(reception.StatusClosed, receptionID).
		WillReturnError(errors.New("ошибка обновления статуса"))

	closed, err := repo.CloseReception(ctx, receptionID)
	require.Error(t, err)
	assert.Nil(t, closed)
}
