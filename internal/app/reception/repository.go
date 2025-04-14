package reception

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"pvz_service/internal/storage/postgres"
)

// mockgen -source=internal/app/reception/repository.go -destination=internal/app/reception/mocks/repository_mock.go -package=mocks

type RepoInterface interface {
	CreateReception(ctx context.Context, reception *Reception) (*Reception, error)
	CheckValidPVZID(ctx context.Context, pvzID uuid.UUID) (bool, error)
	CheckOpenReception(ctx context.Context, pvzID uuid.UUID) (bool, uuid.UUID, error)
	CloseReception(ctx context.Context, receptionID uuid.UUID) (*Reception, error)
}

type Repo struct {
	storage *postgres.Storage
}

func NewReceptionRepository(storage *postgres.Storage) *Repo {
	return &Repo{
		storage: storage,
	}
}

func (r *Repo) CreateReception(ctx context.Context, reception *Reception) (*Reception, error) {
	query := `
			INSERT INTO reception (id, pvz_id, status)
			VALUES ($1, $2, $3)
			RETURNING id, pvz_id, datetime, status;
	`

	var openReception Reception
	err := r.storage.DB.QueryRowContext(
		ctx,
		query,
		reception.ID,
		reception.PvzID,
		StatusInProgress,
	).Scan(
		&openReception.ID,
		&openReception.PvzID,
		&openReception.DateTime,
		&openReception.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания приемки: %w", err)
	}

	return &openReception, nil

}

func (r *Repo) CheckValidPVZID(ctx context.Context, pvzID uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM pvz
		WHERE id = $1
	`

	var count int
	err := r.storage.DB.QueryRowContext(
		ctx,
		query,
		pvzID,
	).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return count > 0, nil
}

func (r *Repo) CheckOpenReception(ctx context.Context, pvzID uuid.UUID) (bool, uuid.UUID, error) {
	query := `
		SELECT id
		FROM reception
		WHERE pvz_id = $1 and status = $2
	`

	var receptionID uuid.UUID
	err := r.storage.DB.QueryRowContext(
		ctx,
		query,
		pvzID,
		StatusInProgress,
	).Scan(&receptionID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, uuid.Nil, nil
		}
		return false, uuid.Nil, err
	}

	return true, receptionID, nil
}

func (r *Repo) CloseReception(ctx context.Context, receptionID uuid.UUID) (*Reception, error) {
	query := `
			UPDATE reception 
			SET status = $1
			WHERE id = $2
			RETURNING id, pvz_id, datetime, status;
	`

	var openReception Reception
	err := r.storage.DB.QueryRowContext(
		ctx,
		query,
		StatusClosed,
		receptionID,
	).Scan(
		&openReception.ID,
		&openReception.PvzID,
		&openReception.DateTime,
		&openReception.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка закрытия приемки: %w", err)
	}

	return &openReception, nil
}
