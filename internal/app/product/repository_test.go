package product_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pvz_service/internal/app/product"
	"pvz_service/internal/storage/postgres"
)

func TestAddProduct_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := product.NewProductRepository(&postgres.Storage{DB: db})

	productID := uuid.New()
	receptionID := uuid.New()
	myProduct := &product.Product{
		ID:          productID,
		Type:        product.TypeElectronics,
		ReceptionID: receptionID,
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO product(id, type, reception_id) 
			VALUES ($1, $2, $3) 
			RETURNING id, datetime, type, reception_id`),
	).
		WithArgs(myProduct.ID, myProduct.Type, myProduct.ReceptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "datetime", "type", "reception_id"}).
			AddRow(myProduct.ID, myProduct.DateTime, myProduct.Type, myProduct.ReceptionID))

	createdProduct, err := repo.AddProduct(context.Background(), myProduct)
	assert.NoError(t, err)
	assert.NotNil(t, createdProduct)
	assert.Equal(t, myProduct.ID, createdProduct.ID)
	assert.Equal(t, myProduct.Type, createdProduct.Type)
	assert.Equal(t, myProduct.ReceptionID, createdProduct.ReceptionID)
}

func TestAddProduct_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := product.NewProductRepository(&postgres.Storage{DB: db})

	productID := uuid.New()
	receptionID := uuid.New()
	product := &product.Product{
		ID:          productID,
		Type:        product.TypeElectronics,
		ReceptionID: receptionID,
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO product(id, type, reception_id) 
			VALUES ($1, $2, $3) 
			RETURNING id, datetime, type, reception_id`),
	).
		WithArgs(product.ID, product.Type, product.ReceptionID).
		WillReturnError(fmt.Errorf("ошибка добавления товара"))

	createdProduct, err := repo.AddProduct(context.Background(), product)
	assert.Error(t, err)
	assert.Nil(t, createdProduct)
	assert.Contains(t, err.Error(), "ошибка добавления товаров")
}

func TestDeleteProduct_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := product.NewProductRepository(&postgres.Storage{DB: db})

	receptionID := uuid.New()
	expectedProduct := &product.Product{
		ID:          uuid.New(),
		Type:        product.TypeClothing,
		ReceptionID: receptionID,
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`DELETE FROM product 
       		WHERE id = (
       			SELECT id 
       			FROM product 
       			WHERE reception_id = $1 
       			ORDER BY datetime 
       			DESC LIMIT 1
			) RETURNING id, datetime, type, reception_id`)).
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "datetime", "type", "reception_id"}).
			AddRow(expectedProduct.ID, expectedProduct.DateTime, expectedProduct.Type, expectedProduct.ReceptionID))

	deletedProduct, err := repo.DeleteProduct(context.Background(), receptionID)
	assert.NoError(t, err)
	assert.NotNil(t, deletedProduct)
	assert.Equal(t, expectedProduct.ID, deletedProduct.ID)
	assert.Equal(t, expectedProduct.Type, deletedProduct.Type)
	assert.Equal(t, expectedProduct.ReceptionID, deletedProduct.ReceptionID)
}

func TestDeleteProduct_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := product.NewProductRepository(&postgres.Storage{DB: db})

	receptionID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(
		`DELETE FROM product 
       		WHERE id = (
       			SELECT id 
       			FROM product 
       			WHERE reception_id = $1 
       			ORDER BY datetime 
       			DESC LIMIT 1)
			RETURNING id, datetime, type, reception_id`),
	).
		WithArgs(receptionID).
		WillReturnError(fmt.Errorf("ошибка удаления товара"))

	deletedProduct, err := repo.DeleteProduct(context.Background(), receptionID)
	assert.Error(t, err)
	assert.Nil(t, deletedProduct)
	assert.Contains(t, err.Error(), "ошибка добавления товаров")
}
