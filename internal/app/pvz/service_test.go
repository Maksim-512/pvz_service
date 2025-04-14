package pvz_test

import (
	"context"
	"errors"
	"go.uber.org/mock/gomock"
	"log/slog"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pvz_service/internal/app/pvz"
	"pvz_service/internal/app/pvz/mocks"
	mockJWT "pvz_service/pkg/jwt/mocks"
)

func newServiceWithMocks(t *testing.T) (*pvz.Service, *mocks.MockRepository, *mockJWT.MockTokenService) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	pvzRepoMock := mocks.NewMockRepository(ctrl)
	logger := slog.Default()
	mockToken := mockJWT.NewMockTokenService(ctrl)

	service := pvz.NewPVZService(logger, pvzRepoMock)

	return service, pvzRepoMock, mockToken
}

func contextWithReqID() context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, middleware.RequestIDKey, "tests-request-id")
}

func TestServiceAddPVZ_Success(t *testing.T) {
	service, repo, _ := newServiceWithMocks(t)
	ctx := contextWithReqID()

	req := &pvz.CreatePVZRequest{City: "Москва"}
	expected := &pvz.PVZ{ID: uuid.New(), City: "Москва"}

	repo.EXPECT().CreatePVZ(gomock.Any(), gomock.Any()).Return(expected, nil)

	result, err := service.ServiceAddPVZ(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, expected.City, result.City)
}

func TestServiceAddPVZ_InvalidCity(t *testing.T) {
	service, _, _ := newServiceWithMocks(t)
	ctx := contextWithReqID()

	req := &pvz.CreatePVZRequest{City: "Мытищи"}

	result, err := service.ServiceAddPVZ(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "недопустимый город: Мытищи")
}

func TestServiceAddPVZ_RepoError(t *testing.T) {
	service, repo, _ := newServiceWithMocks(t)
	ctx := contextWithReqID()

	req := &pvz.CreatePVZRequest{City: "Казань"}

	repo.EXPECT().CreatePVZ(gomock.Any(), gomock.Any()).Return(nil, errors.New("ошибка БД"))

	result, err := service.ServiceAddPVZ(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "ошибка БД")
}

func TestServiceGetPVZs_Success(t *testing.T) {
	service, repo, _ := newServiceWithMocks(t)
	ctx := contextWithReqID()

	startStr := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	endStr := time.Now().Format(time.RFC3339)

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	page := "1"
	limit := "10"

	expected := []*pvz.PVZWithReceptions{
		{PVZ: &pvz.PVZ{City: "Москва"}},
	}

	repo.EXPECT().GetPVZWithReceptions(ctx, &start, &end, 1, 10).Return(expected, nil)

	result, err := service.ServiceGetPVZs(ctx, startStr, endStr, page, limit)

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Москва", result[0].PVZ.City)
}

func TestServiceGetPVZs_InvalidPage(t *testing.T) {
	service, _, _ := newServiceWithMocks(t)
	ctx := contextWithReqID()

	result, err := service.ServiceGetPVZs(ctx, "", "", "abc", "10")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestServiceGetPVZs_InvalidLimit(t *testing.T) {
	service, _, _ := newServiceWithMocks(t)
	ctx := contextWithReqID()

	result, err := service.ServiceGetPVZs(ctx, "", "", "1", "100")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "limit должен быть от 1 до 30 включительно")
}

func TestServiceGetPVZs_InvalidStartDate(t *testing.T) {
	service, _, _ := newServiceWithMocks(t)
	ctx := contextWithReqID()

	result, err := service.ServiceGetPVZs(ctx, "bad-date", "", "1", "10")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "неверный формат startDate, должен быть RFC3339")
}

func TestServiceGetPVZs_InvalidEndDate(t *testing.T) {
	service, _, _ := newServiceWithMocks(t)
	ctx := contextWithReqID()

	start := time.Now().Format(time.RFC3339)
	result, err := service.ServiceGetPVZs(ctx, start, "oops", "1", "10")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "неверный формат endDate, должен быть RFC3339")
}

func TestServiceGetPVZs_RepoError(t *testing.T) {
	service, repo, _ := newServiceWithMocks(t)
	ctx := contextWithReqID()

	startStr := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	endStr := time.Now().Format(time.RFC3339)

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	repo.EXPECT().GetPVZWithReceptions(
		ctx,
		&start,
		&end,
		1,
		10,
	).Return(nil, errors.New("ошибка БД"))

	result, err := service.ServiceGetPVZs(ctx, startStr, endStr, "1", "10")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "ошибка БД")
}

func TestValidateParamsPage_Empty(t *testing.T) {
	page, err := pvz.ValidateParamsPage("")
	assert.NoError(t, err)
	assert.Equal(t, 1, page)
}

func TestValidateParamsPage_NotANumber(t *testing.T) {
	_, err := pvz.ValidateParamsPage("abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid syntax")
}

func TestValidateParamsPage_Zero(t *testing.T) {
	_, err := pvz.ValidateParamsPage("0")
	assert.Error(t, err)
	assert.Equal(t, "page должен быть ≥ 1", err.Error())
}

func TestValidateParamsPage_Negative(t *testing.T) {
	_, err := pvz.ValidateParamsPage("-2")
	assert.Error(t, err)
	assert.Equal(t, "page должен быть ≥ 1", err.Error())
}

func TestValidateParamsPage_Valid(t *testing.T) {
	page, err := pvz.ValidateParamsPage("5")
	assert.NoError(t, err)
	assert.Equal(t, 5, page)
}

func TestValidateParamsLimit_Empty(t *testing.T) {
	limit, err := pvz.ValidateParamsLimit("")
	assert.NoError(t, err)
	assert.Equal(t, 10, limit)
}

func TestValidateParamsLimit_NotANumber(t *testing.T) {
	_, err := pvz.ValidateParamsLimit("abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "limit должен быть числом")
}

func TestValidateParamsLimit_Zero(t *testing.T) {
	_, err := pvz.ValidateParamsLimit("0")
	assert.Error(t, err)
	assert.Equal(t, "limit должен быть от 1 до 30 включительно", err.Error())
}

func TestValidateParamsLimit_TooBig(t *testing.T) {
	_, err := pvz.ValidateParamsLimit("50")
	assert.Error(t, err)
	assert.Equal(t, "limit должен быть от 1 до 30 включительно", err.Error())
}

func TestValidateParamsLimit_Valid(t *testing.T) {
	limit, err := pvz.ValidateParamsLimit("20")
	assert.NoError(t, err)
	assert.Equal(t, 20, limit)
}

func TestValidateDate_Empty(t *testing.T) {
	date, err := pvz.ValidateDate("")
	assert.NoError(t, err)
	assert.Nil(t, date)
}
