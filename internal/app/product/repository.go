package product

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"pvz_service/internal/storage/postgres"
)

// mockgen -source=internal/app/product/repository.go -destination=internal/app/product/mocks/repository_mock.go -package=mocks

type RepoInterface interface {
	AddProduct(ctx context.Context, product *Product) (*Product, error)
	DeleteProduct(ctx context.Context, receptionID uuid.UUID) (*Product, error)
}

type Repo struct {
	storage *postgres.Storage
}

func NewProductRepository(storage *postgres.Storage) *Repo {
	return &Repo{
		storage: storage,
	}
}

func (p *Repo) AddProduct(ctx context.Context, product *Product) (*Product, error) {
	query := `
			INSERT INTO product(id, type, reception_id)
			VALUES ($1, $2, $3)
			RETURNING id, datetime, type, reception_id
			`

	var createdProduct Product
	err := p.storage.DB.QueryRowContext(
		ctx,
		query,
		product.ID,
		product.Type,
		product.ReceptionID,
	).Scan(
		&createdProduct.ID,
		&createdProduct.DateTime,
		&createdProduct.Type,
		&createdProduct.ReceptionID,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка добавления товаров: %w", err)
	}

	return &createdProduct, nil
}

func (p *Repo) DeleteProduct(ctx context.Context, receptionID uuid.UUID) (*Product, error) {
	query := `
			DELETE FROM product
			WHERE id = (
				SELECT id 
				FROM product
				WHERE reception_id = $1
				ORDER BY datetime DESC
				LIMIT 1
			)
			RETURNING id, datetime, type, reception_id
			`

	var createdProduct Product
	err := p.storage.DB.QueryRowContext(
		ctx,
		query,
		receptionID,
	).Scan(
		&createdProduct.ID,
		&createdProduct.DateTime,
		&createdProduct.Type,
		&createdProduct.ReceptionID,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка добавления товаров: %w", err)
	}

	return &createdProduct, nil
}
