package pvz_test

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

	"pvz_service/internal/app/pvz"
	"pvz_service/internal/storage/postgres"
)

func TestCreatePVZ_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	pvzID := uuid.New()
	city := "Москва"
	pvzToCreate := &pvz.PVZ{
		ID:   pvzID,
		City: city,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO pvz (id, city)
		VALUES ($1, $2)
		RETURNING id, city, registration_date;
	`)).
		WithArgs(pvzID, city).
		WillReturnRows(sqlmock.NewRows([]string{"id", "city", "registration_date"}).
			AddRow(pvzID, city, time.Now()))

	createdPVZ, err := repo.CreatePVZ(context.Background(), pvzToCreate)

	assert.NoError(t, err)
	assert.NotNil(t, createdPVZ)
	assert.Equal(t, pvzID, createdPVZ.ID)
	assert.Equal(t, city, createdPVZ.City)
}

func TestCreatePVZ_Failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	pvzID := uuid.New()
	city := "Москва"
	pvzToCreate := &pvz.PVZ{
		ID:   pvzID,
		City: city,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO pvz (id, city)
		VALUES ($1, $2)
		RETURNING id, city, registration_date;
	`)).
		WithArgs(pvzID, city).
		WillReturnError(fmt.Errorf("ошибка создания ПВЗ"))

	createdPVZ, err := repo.CreatePVZ(context.Background(), pvzToCreate)

	assert.Error(t, err)
	assert.Nil(t, createdPVZ)
}

func TestGetPVZWithReceptions_Failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	startDate := time.Now().Add(-time.Hour * 24)
	endDate := time.Now()
	page, limit := 1, 10

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT gen.id, gen.registration_date, gen.city
		FROM gen
		JOIN reception ON gen.id = reception.pvz_id
		WHERE ($1::timestamp IS NULL OR reception.datetime >= $1)
		  AND ($2::timestamp IS NULL OR reception.datetime <= $2)
		GROUP BY gen.id
		ORDER BY gen.city
		OFFSET $3 LIMIT $4
	`)).
		WithArgs(startDate, endDate, (page-1)*limit, limit).
		WillReturnError(fmt.Errorf("ошибка получения ПВЗ"))

	getPVZ, err := repo.GetPVZWithReceptions(context.Background(), &startDate, &endDate, page, limit)

	assert.Error(t, err)
	assert.Nil(t, getPVZ)
}

func TestGetReceptionsWithProducts_Failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	pvzID := uuid.New()
	startDate := time.Now().Add(-time.Hour * 24)
	endDate := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, datetime, status
		FROM reception
		WHERE pvz_id = $1
		AND ($2::timestamp IS NULL OR datetime >= $2)
		AND ($3::timestamp IS NULL OR datetime <= $3)
		ORDER BY datetime DESC
	`)).
		WithArgs(pvzID, startDate, endDate).
		WillReturnError(fmt.Errorf("ошибка получения получений"))

	receptions, err := repo.GetReceptionsWithProducts(context.Background(), pvzID, &startDate, &endDate)

	assert.Error(t, err)
	assert.Nil(t, receptions)
}

func TestGetProductsByReception_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	receptionID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, datetime, type, reception_id
		FROM product
		WHERE reception_id = $1
		ORDER BY datetime DESC
	`)).
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "datetime", "type", "reception_id"}).
			AddRow(uuid.New(), time.Now(), "type1", receptionID).
			AddRow(uuid.New(), time.Now().Add(-time.Hour), "type2", receptionID))

	products, err := repo.GetProductsByReception(context.Background(), receptionID)

	assert.NoError(t, err)
	assert.NotNil(t, products)
	assert.Len(t, products, 2)
	assert.Equal(t, "type1", products[0].Type)
	assert.Equal(t, "type2", products[1].Type)
}

func TestGetProductsByReception_Failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	receptionID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, datetime, type, reception_id
		FROM product
		WHERE reception_id = $1
		ORDER BY datetime DESC
	`)).
		WithArgs(receptionID).
		WillReturnError(fmt.Errorf("ошибка получения продуктов"))

	products, err := repo.GetProductsByReception(context.Background(), receptionID)

	assert.Error(t, err)
	assert.Nil(t, products)
}

func TestGetReceptionsWithProducts_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	pvzID := uuid.New()
	startDate := time.Now().Add(-time.Hour * 24)
	endDate := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, datetime, status
		FROM reception
		WHERE pvz_id = $1
		AND ($2::timestamp IS NULL OR datetime >= $2)
		AND ($3::timestamp IS NULL OR datetime <= $3)
		ORDER BY datetime DESC
	`)).
		WithArgs(pvzID, startDate, endDate).
		WillReturnRows(sqlmock.NewRows([]string{"id", "datetime", "status"}).
			AddRow(uuid.New(), time.Now(), "received").
			AddRow(uuid.New(), time.Now().Add(-time.Hour), "processing"))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, datetime, type, reception_id
		FROM product
		WHERE reception_id = $1
		ORDER BY datetime DESC
	`)).
		WithArgs(uuid.New()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "datetime", "type", "reception_id"}).
			AddRow(uuid.New(), time.Now(), "type1", uuid.New()).
			AddRow(uuid.New(), time.Now().Add(-time.Hour), "type2", uuid.New())).
		WillReturnError(fmt.Errorf("ошибка при сканировании продуктов"))

	receptions, err := repo.GetReceptionsWithProducts(context.Background(), pvzID, &startDate, &endDate)

	assert.Error(t, err)
	assert.Nil(t, receptions)
}

func TestGetPVZWithReceptions_CloseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	startDate := time.Now().Add(-time.Hour * 24)
	endDate := time.Now()
	page, limit := 1, 10

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT gen.id, gen.registration_date, gen.city
		FROM gen
		JOIN reception ON gen.id = reception.pvz_id
		WHERE ($1::timestamp IS NULL OR reception.datetime >= $1)
		  AND ($2::timestamp IS NULL OR reception.datetime <= $2)
		GROUP BY gen.id
		ORDER BY gen.city
		OFFSET $3 LIMIT $4
	`)).
		WithArgs(startDate, endDate, (page-1)*limit, limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "city", "registration_date"}).
			AddRow(uuid.New(), "Москва", time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT gen.id, gen.registration_date, gen.city
		FROM gen
		JOIN reception ON gen.id = reception.pvz_id
		WHERE ($1::timestamp IS NULL OR reception.datetime >= $1)
		  AND ($2::timestamp IS NULL OR reception.datetime <= $2)
		GROUP BY gen.id
		ORDER BY gen.city
		OFFSET $3 LIMIT $4
	`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "city", "registration_date"}).
			AddRow(uuid.New(), "Москва", time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT gen.id, gen.registration_date, gen.city
		FROM gen
		JOIN reception ON gen.id = reception.pvz_id
		WHERE ($1::timestamp IS NULL OR reception.datetime >= $1)
		  AND ($2::timestamp IS NULL OR reception.datetime <= $2)
		GROUP BY gen.id
		ORDER BY gen.city
		OFFSET $3 LIMIT $4
	`)).
		WillReturnError(fmt.Errorf("ошибка при закрытии rows"))

	getPVZ, err := repo.GetPVZWithReceptions(context.Background(), &startDate, &endDate, page, limit)

	assert.Error(t, err)
	assert.Nil(t, getPVZ)
}

func TestGetPVZWithReceptions_RowsScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	startDate := time.Now().Add(-time.Hour * 24)
	endDate := time.Now()
	page, limit := 1, 10

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT gen.id, gen.registration_date, gen.city
        FROM gen
        JOIN reception ON gen.id = reception.pvz_id
        WHERE ($1::timestamp IS NULL OR reception.datetime >= $1)
          AND ($2::timestamp IS NULL OR reception.datetime <= $2)
        GROUP BY gen.id
        ORDER BY gen.city
        OFFSET $3 LIMIT $4
    `)).
		WithArgs(startDate, endDate, (page-1)*limit, limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date"}).
			AddRow(uuid.New(), time.Now()))

	getPVZ, err := repo.GetPVZWithReceptions(context.Background(), &startDate, &endDate, page, limit)

	assert.Error(t, err)
	assert.Nil(t, getPVZ)
}

func TestGetPVZWithReceptions_ReceptionsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	startDate := time.Now().Add(-time.Hour * 24)
	endDate := time.Now()
	page, limit := 1, 10
	pvzID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT gen.id, gen.registration_date, gen.city
        FROM gen
        JOIN reception ON gen.id = reception.pvz_id
        WHERE ($1::timestamp IS NULL OR reception.datetime >= $1)
          AND ($2::timestamp IS NULL OR reception.datetime <= $2)
        GROUP BY gen.id
        ORDER BY gen.city
        OFFSET $3 LIMIT $4
    `)).
		WithArgs(startDate, endDate, (page-1)*limit, limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow(pvzID, time.Now(), "Москва"))

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, datetime, status
        FROM reception
        WHERE pvz_id = $1
        AND ($2::timestamp IS NULL OR datetime >= $2)
        AND ($3::timestamp IS NULL OR datetime <= $3)
        ORDER BY datetime DESC
    `)).
		WithArgs(pvzID, startDate, endDate).
		WillReturnError(fmt.Errorf("ошибка получения приемок"))

	getPVZ, err := repo.GetPVZWithReceptions(context.Background(), &startDate, &endDate, page, limit)

	assert.Error(t, err)
	assert.Nil(t, getPVZ)
}

func TestGetReceptionsWithProducts_RowsScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	pvzID := uuid.New()
	startDate := time.Now().Add(-time.Hour * 24)
	endDate := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, datetime, status
        FROM reception
        WHERE pvz_id = $1
        AND ($2::timestamp IS NULL OR datetime >= $2)
        AND ($3::timestamp IS NULL OR datetime <= $3)
        ORDER BY datetime DESC
    `)).
		WithArgs(pvzID, startDate, endDate).
		WillReturnRows(sqlmock.NewRows([]string{"id", "datetime"}).
			AddRow(uuid.New(), time.Now()))

	receptions, err := repo.GetReceptionsWithProducts(context.Background(), pvzID, &startDate, &endDate)

	assert.Error(t, err)
	assert.Nil(t, receptions)
}

func TestGetProductsByReception_RowsScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	receptionID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, datetime, type, reception_id
        FROM product
        WHERE reception_id = $1
        ORDER BY datetime DESC
    `)).
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "datetime", "reception_id"}).
			AddRow(uuid.New(), time.Now(), receptionID))

	products, err := repo.GetProductsByReception(context.Background(), receptionID)

	assert.Error(t, err)
	assert.Nil(t, products)
}

func TestGetPVZWithReceptions_SuccessWithEmptyReceptions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	startDate := time.Now().Add(-time.Hour * 24)
	endDate := time.Now()
	page, limit := 1, 10
	pvzID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT pvz.id, pvz.registration_date, pvz.city
        FROM pvz
        JOIN reception ON pvz.id = reception.pvz_id
        WHERE ($1::timestamp IS NULL OR reception.datetime >= $1)
          AND ($2::timestamp IS NULL OR reception.datetime <= $2)
        GROUP BY pvz.id
        ORDER BY pvz.city
        OFFSET $3 LIMIT $4
    `)).
		WithArgs(startDate, endDate, (page-1)*limit, limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow(pvzID, time.Now(), "Москва"))

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, datetime, status
        FROM reception
        WHERE pvz_id = $1
        AND ($2::timestamp IS NULL OR datetime >= $2)
        AND ($3::timestamp IS NULL OR datetime <= $3)
        ORDER BY datetime DESC
    `)).
		WithArgs(pvzID, startDate, endDate).
		WillReturnRows(sqlmock.NewRows([]string{"id", "datetime", "status"}))

	getPVZ, err := repo.GetPVZWithReceptions(context.Background(), &startDate, &endDate, page, limit)

	assert.NoError(t, err)
	require.NotNil(t, getPVZ)
	require.Len(t, getPVZ, 1)
	assert.Equal(t, pvzID, getPVZ[0].PVZ.ID)
	assert.Empty(t, getPVZ[0].Receptions)
}

func TestGetReceptionsWithProducts_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := pvz.NewPVZRepository(&postgres.Storage{DB: db})

	pvzID := uuid.New()
	receptionID := uuid.New()
	startDate := time.Now().Add(-time.Hour * 24)
	endDate := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, datetime, status
		FROM reception
		WHERE pvz_id = $1
		AND ($2::timestamp IS NULL OR datetime >= $2)
		AND ($3::timestamp IS NULL OR datetime <= $3)
		ORDER BY datetime DESC
	`)).
		WithArgs(pvzID, startDate, endDate).
		WillReturnRows(sqlmock.NewRows([]string{"id", "datetime", "status"}).
			AddRow(receptionID, time.Now(), "received"))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, datetime, type, reception_id
		FROM product
		WHERE reception_id = $1
		ORDER BY datetime DESC
	`)).
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "datetime", "type", "reception_id"}).
			AddRow(uuid.New(), time.Now(), "type1", receptionID).
			AddRow(uuid.New(), time.Now().Add(-time.Hour), "type2", receptionID))

	receptions, err := repo.GetReceptionsWithProducts(context.Background(), pvzID, &startDate, &endDate)

	assert.NoError(t, err)
	assert.NotNil(t, receptions)
	assert.Len(t, receptions, 1)
	assert.Equal(t, "received", receptions[0].Reception.Status)
	assert.Len(t, receptions[0].Products, 2)
	assert.Equal(t, "type1", receptions[0].Products[0].Type)
	assert.Equal(t, "type2", receptions[0].Products[1].Type)
}
