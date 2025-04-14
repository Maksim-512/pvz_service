package pvz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"pvz_service/internal/app/auth"
	"pvz_service/internal/app/product"
	"pvz_service/internal/app/pvz"
	"pvz_service/internal/app/reception"
	"pvz_service/internal/config"
	"pvz_service/internal/lib/logger"
	"pvz_service/internal/server"
	"pvz_service/internal/storage/postgres"
	"pvz_service/pkg/jwt"
)

type TestServer struct {
	Handler http.Handler
	T       *testing.T
	DB      *postgres.Storage
}

func NewTestServer(t *testing.T) *TestServer {
	cfg, err := config.Load("../../config/config.local.yaml")
	require.NoError(t, err)

	jwtService := jwt.NewJWTService(cfg.Auth.SecretKey)

	myLogger := logger.SetupLogger(cfg.Logging.Env)

	storage, err := postgres.New(&cfg.Database)
	require.NoError(t, err)

	err = applyMigrations(storage)
	require.NoError(t, err)

	ctx := context.Background()
	err = storage.Ping(ctx)
	require.NoError(t, err)

	userRepo := auth.NewUserRepository(storage)
	authService := auth.NewAuthService(myLogger, userRepo, jwtService)
	authHandler := auth.NewAuthHandler(myLogger, authService)

	pvzRepo := pvz.NewPVZRepository(storage)
	pvzService := pvz.NewPVZService(myLogger, pvzRepo)
	pvzHandler := pvz.NewPVZHandler(myLogger, pvzService)

	receptionRepo := reception.NewReceptionRepository(storage)
	receptionService := reception.NewReceptionService(myLogger, receptionRepo)
	receptionHandler := reception.NewReceptionHandler(myLogger, receptionService)

	productRepo := product.NewProductRepository(storage)
	productService := product.NewProductService(myLogger, productRepo, receptionRepo)
	productHandler := product.NewProductHandler(myLogger, productService)

	router := server.NewRouter(
		myLogger,
		jwtService,
		authHandler,
		pvzHandler,
		receptionHandler,
		productHandler)

	return &TestServer{
		Handler: router,
		T:       t,
		DB:      storage,
	}
}

func (ts *TestServer) Close() {
	ts.DB.Close()
}

func (s *TestServer) Do(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Handler.ServeHTTP(rr, req)
	return rr
}

func TestCreatePvzAndReceiveAsModeratorAndEmployee(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	moderatorToken := dummyLogin(ts, auth.RoleModerator)

	pvzResp := createPVZ(ts, moderatorToken, "Москва")
	fmt.Println("pvzResp: ", pvzResp)

	require.Equal(t, pvzResp.City, "Москва")

	employeeToken := dummyLogin(ts, auth.RoleEmployee)

	receiveResp := createReceptions(ts, employeeToken, pvzResp)
	require.Equal(t, receiveResp.PvzID, pvzResp.ID)
	require.Equal(t, receiveResp.Status, reception.StatusInProgress)

	productTypes := []string{product.TypeElectronics, product.TypeClothing, product.TypeShoes}
	for i := 0; i < 50; i++ {
		productType := productTypes[i%3]
		productResp := addProduct(ts, employeeToken, pvzResp.ID, productType)
		require.Equal(t, productType, productResp.Type)
		require.Equal(t, receiveResp.ID, productResp.ReceptionID)
	}

	productsCount := countProductsInReception(ts, receiveResp.ID)
	require.Equal(t, 50, productsCount)

	closedReception := closeReception(ts, employeeToken, pvzResp.ID)
	require.Equal(t, receiveResp.ID, closedReception.ID)
	require.Equal(t, reception.StatusClosed, closedReception.Status)

	err := tryAddProductToClosedReception(ts, employeeToken, pvzResp.ID)
	require.NoError(t, err)
}

func applyMigrations(db *postgres.Storage) error {
	migrationFile := "../../migrations/init.sql"

	sqlBytes, err := os.ReadFile(migrationFile)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.DB.ExecContext(ctx, string(sqlBytes))
	return err
}

func dummyLogin(ts *TestServer, role string) *auth.LoginResponse {
	ts.T.Log("Отправка запроса на /dummyLogin с ролью:", role)

	body := auth.DummyLoginRequest{
		Role: role,
	}

	bodyData, err := json.Marshal(body)
	if err != nil {
		ts.T.Fatal("Ошибка сериализации запроса:", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(bodyData))
	req.Header.Set("Content-Type", "application/json")

	resp := ts.Do(req)
	require.Equal(ts.T, http.StatusOK, resp.Code)

	var token auth.LoginResponse
	err = json.Unmarshal(resp.Body.Bytes(), &token)
	require.NoError(ts.T, err)

	return &token
}

func createPVZ(ts *TestServer, token *auth.LoginResponse, city string) *pvz.PVZ {
	ts.T.Log("Создание нового PVZ для города:", city)

	createRequest := &pvz.CreatePVZRequest{
		City: city,
	}
	pvzData, err := json.Marshal(createRequest)
	require.NoError(ts.T, err)

	req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewBuffer(pvzData))
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp := ts.Do(req)
	require.Equal(ts.T, http.StatusCreated, resp.Code)

	var pvzResp *pvz.PVZ
	err = json.Unmarshal(resp.Body.Bytes(), &pvzResp)
	require.NoError(ts.T, err)

	return pvzResp
}

func createReceptions(ts *TestServer, token *auth.LoginResponse, pvz *pvz.PVZ) *reception.Reception {
	ts.T.Log("Создание новой приемки заказа:", pvz)

	receiveRequest := &reception.OpenReceptionRequest{
		PvzID: pvz.ID,
	}
	receiveData, err := json.Marshal(receiveRequest)
	require.NoError(ts.T, err)

	req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer(receiveData))
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp := ts.Do(req)
	require.Equal(ts.T, http.StatusCreated, resp.Code)

	var receiveResp *reception.Reception
	err = json.Unmarshal(resp.Body.Bytes(), &receiveResp)
	require.NoError(ts.T, err)

	return receiveResp
}

func addProduct(ts *TestServer, token *auth.LoginResponse, pvzID uuid.UUID, productType string) *product.Product {
	ts.T.Log("Добавление товара типа:", productType)

	productReq := product.CreateProductRequest{
		Type:  productType,
		PvzID: pvzID,
	}
	productData, err := json.Marshal(productReq)
	require.NoError(ts.T, err)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(productData))
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp := ts.Do(req)
	require.Equal(ts.T, http.StatusCreated, resp.Code)

	var productResp *product.Product
	err = json.Unmarshal(resp.Body.Bytes(), &productResp)
	require.NoError(ts.T, err)

	return productResp
}

func closeReception(ts *TestServer, token *auth.LoginResponse, pvzID uuid.UUID) *reception.Reception {
	ts.T.Log("Закрытие приемки для PVZ:", pvzID)

	url := fmt.Sprintf("/pvz/%s/close_last_reception", pvzID.String())
	req := httptest.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp := ts.Do(req)
	require.Equal(ts.T, http.StatusOK, resp.Code)

	var receptionResp *reception.Reception
	err := json.Unmarshal(resp.Body.Bytes(), &receptionResp)
	require.NoError(ts.T, err)

	return receptionResp
}

func countProductsInReception(ts *TestServer, receptionID uuid.UUID) int {
	var count int
	err := ts.DB.DB.QueryRow(
		"SELECT COUNT(*) FROM product WHERE reception_id = $1",
		receptionID,
	).Scan(&count)
	require.NoError(ts.T, err)
	return count
}

func tryAddProductToClosedReception(ts *TestServer, token *auth.LoginResponse, pvzID uuid.UUID) error {
	ts.T.Log("Попытка добавить товар в закрытую приемку")

	productReq := product.CreateProductRequest{
		Type:  product.TypeElectronics,
		PvzID: pvzID,
	}
	productData, err := json.Marshal(productReq)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга запроса: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(productData))
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp := ts.Do(req)

	if resp.Code != http.StatusBadRequest {
		return fmt.Errorf("ожидался статус 400, получен %d", resp.Code)
	}

	var errorResp struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &errorResp); err != nil {
		return fmt.Errorf("ошибка декодирования ответа: %v", err)
	}

	if errorResp.Message == "" {
		return fmt.Errorf("ожидалось сообщение об ошибке, получен пустой ответ")
	}

	return nil
}
