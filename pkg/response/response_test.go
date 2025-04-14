package response_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"pvz_service/pkg/response"
)

func TestMyResponseError(t *testing.T) {

	rr := httptest.NewRecorder()

	response.MyResponseError(rr, http.StatusBadRequest, "Ошибка запроса")

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	expected := `{"message":"Ошибка запроса"}`
	assert.JSONEq(t, expected, rr.Body.String())
}

func TestMyResponseJSON(t *testing.T) {
	type responseData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := responseData{
		Name:  "tests",
		Value: 111,
	}
	rr := httptest.NewRecorder()

	response.MyResponseJSON(rr, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, rr.Code)

	expected := `{"name":"tests","value":111}`
	assert.JSONEq(t, expected, rr.Body.String())
}
