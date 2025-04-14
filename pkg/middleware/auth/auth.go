package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"pvz_service/pkg/jwt"
	"pvz_service/pkg/response"
)

func CheckAuthMiddleware(jwtService jwt.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			tokenString := r.Header.Get("Authorization")

			if tokenString == "" {
				log.Printf("Отсутствует токен авторизации")
				response.MyResponseError(w, http.StatusForbidden, "Доступ запрещен")
				return
			}

			tokenString = strings.TrimPrefix(tokenString, "Bearer ")

			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				log.Printf("Некорректный token: %v", err)
				response.MyResponseError(w, http.StatusForbidden, "Доступ запрещен")
				return
			}

			role, ok := claims["role"].(string)
			if !ok {
				log.Printf("Неверная роль: %s", role)
				response.MyResponseError(w, http.StatusForbidden, "Доступ запрещен")
				return
			}

			ctx := context.WithValue(r.Context(), "role", role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleMiddle(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleValue := r.Context().Value("role")
			role, ok := roleValue.(string)
			if !ok {
				response.MyResponseError(w, http.StatusForbidden, "Доступ запрещен")
				return
			}

			for _, requiredRole := range requiredRoles {
				if role == requiredRole {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.MyResponseError(w, http.StatusForbidden, "Доступ запрещен")
		})
	}
}
