package jwt

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// mockgen -source=pkg/jwt/jwt.go -destination=pkg/jwt/mocks/jwt_mock.go -package=mocks

type JWTService struct {
	secretKey        string
	SignedStringFunc func(token *jwt.Token, secretKey string) (string, error)
}

func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey: secretKey,
		SignedStringFunc: func(token *jwt.Token, secretKey string) (string, error) {
			return token.SignedString([]byte(secretKey))
		},
	}
}

type TokenService interface {
	GenerateToken(role string) (*TokenResponse, error)
	ValidateToken(tokenString string) (map[string]interface{}, error)
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (j *JWTService) GenerateToken(role string) (*TokenResponse, error) {
	claims := jwt.MapClaims{
		"role": role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour * 1).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := j.SignedStringFunc(token, j.secretKey)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		Token: tokenString,
	}, nil
}

func (j *JWTService) ValidateToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("токен невалидый: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("не удалось извлечь данные из токена")
	}

	return claims, nil
}
