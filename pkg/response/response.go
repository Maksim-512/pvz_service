package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Errors string `json:"message"`
}

func MyResponseError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(ErrorResponse{
		Errors: message,
	})
	if err != nil {
		http.Error(w, "Ошибка шифрования", http.StatusInternalServerError)
		return
	}
}

func MyResponseJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		http.Error(w, "Ошибка шифрования", http.StatusInternalServerError)
	}
}
