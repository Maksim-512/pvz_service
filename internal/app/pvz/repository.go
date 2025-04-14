package pvz

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"pvz_service/internal/app/product"
	"pvz_service/internal/app/reception"
	"pvz_service/internal/storage/postgres"
)

// mockgen -source=internal/app/gen/repository.go -destination=internal/app/gen/mocks/repository_mock.go -package=mocks

type RepoInterface interface {
	CreatePVZ(ctx context.Context, pvz *PVZ) (*PVZ, error)
	GetPVZWithReceptions(
		ctx context.Context,
		startDate,
		endDate *time.Time,
		page,
		limit int,
	) ([]*PVZWithReceptions, error)
}

type Repo struct {
	storage *postgres.Storage
}

func NewPVZRepository(storage *postgres.Storage) *Repo {
	return &Repo{
		storage: storage,
	}
}

func (p *Repo) CreatePVZ(ctx context.Context, pvz *PVZ) (*PVZ, error) {
	query := `
			INSERT INTO pvz (id, city)
			VALUES ($1, $2)
			RETURNING id, city, registration_date;
			`

	var createdPVZ PVZ
	err := p.storage.DB.QueryRowContext(
		ctx,
		query,
		pvz.ID,
		pvz.City,
	).Scan(
		&createdPVZ.ID,
		&createdPVZ.City,
		&createdPVZ.RegistrationDate,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания ПВЗ: %w", err)
	}

	return &createdPVZ, nil
}

func (p *Repo) GetPVZWithReceptions(
	ctx context.Context,
	startDate,
	endDate *time.Time,
	page,
	limit int,
) ([]*PVZWithReceptions, error) {
	offset := (page - 1) * limit

	query := `
		SELECT pvz.id, pvz.registration_date, pvz.city
		FROM pvz
		JOIN reception ON pvz.id = reception.pvz_id
		WHERE ($1::timestamp IS NULL OR reception.datetime >= $1)
		  AND ($2::timestamp IS NULL OR reception.datetime <= $2)
		GROUP BY pvz.id
		ORDER BY pvz.city
		OFFSET $3 LIMIT $4
	`

	rows, err := p.storage.DB.QueryContext(ctx, query, startDate, endDate, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*PVZWithReceptions
	for rows.Next() {
		var pvz PVZ
		if err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City); err != nil {
			return nil, err
		}

		receptions, err := p.GetReceptionsWithProducts(ctx, pvz.ID, startDate, endDate)
		if err != nil {
			return nil, err
		}

		result = append(result, &PVZWithReceptions{
			PVZ:        &pvz,
			Receptions: receptions,
		})
	}

	return result, nil
}

func (p *Repo) GetReceptionsWithProducts(
	ctx context.Context,
	pvzID uuid.UUID,
	startDate,
	endDate *time.Time,
) ([]*ReceptionWithProducts, error) {
	query := `
		SELECT id, datetime, status
		FROM reception
		WHERE pvz_id = $1
		AND ($2::timestamp IS NULL OR datetime >= $2)
		AND ($3::timestamp IS NULL OR datetime <= $3)
		ORDER BY datetime DESC
	`

	rows, err := p.storage.DB.QueryContext(ctx, query, pvzID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*ReceptionWithProducts

	for rows.Next() {
		var rcp reception.Reception
		if err := rows.Scan(&rcp.ID, &rcp.DateTime, &rcp.Status); err != nil {
			return nil, err
		}
		rcp.PvzID = pvzID

		products, err := p.GetProductsByReception(ctx, rcp.ID)
		if err != nil {
			return nil, err
		}

		result = append(result, &ReceptionWithProducts{
			Reception: &rcp,
			Products:  products,
		})
	}

	return result, nil
}

func (p *Repo) GetProductsByReception(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	query := `
		SELECT id, datetime, type, reception_id
		FROM product
		WHERE reception_id = $1
		ORDER BY datetime DESC
	`

	rows, err := p.storage.DB.QueryContext(ctx, query, receptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*product.Product
	for rows.Next() {
		var pr product.Product
		if err := rows.Scan(&pr.ID, &pr.DateTime, &pr.Type, &pr.ReceptionID); err != nil {
			return nil, err
		}
		products = append(products, &pr)
	}

	return products, nil
}
