package jwt_test

import (
	"errors"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pvz_service/internal/app/auth"
	myJWT "pvz_service/pkg/jwt"
)

const secretKey = "testsecret"

func TestGenerateToken(t *testing.T) {
	jwtService := myJWT.NewJWTService(secretKey)

	tokenResponse, err := jwtService.GenerateToken(auth.RoleModerator)
	require.NoError(t, err)
	require.NotNil(t, tokenResponse)

	assert.NotEmpty(t, tokenResponse.Token)

	parsedToken, err := jwt.Parse(tokenResponse.Token, func(_ *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)

	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, auth.RoleModerator, claims["role"])
}

func TestValidateToken_Valid(t *testing.T) {
	jwtService := myJWT.NewJWTService(secretKey)

	tokenResponse, err := jwtService.GenerateToken(auth.RoleEmployee)
	require.NoError(t, err)
	require.NotNil(t, tokenResponse)

	claims, err := jwtService.ValidateToken(tokenResponse.Token)
	require.NoError(t, err)

	assert.Equal(t, auth.RoleEmployee, claims["role"])

	assert.True(t, claims["iat"].(float64) <= float64(time.Now().Unix()))
	assert.True(t, claims["exp"].(float64) > float64(time.Now().Unix()))
}

func TestValidateToken_Invalid(t *testing.T) {
	jwtService := myJWT.NewJWTService(secretKey)

	tokenResponse, err := jwtService.GenerateToken(auth.RoleEmployee)
	require.NoError(t, err)

	invalidToken := tokenResponse.Token + "invalid"

	claims, err := jwtService.ValidateToken(invalidToken)
	require.Error(t, err)
	require.Nil(t, claims)
}

func TestGenerateToken_SignedStringError(t *testing.T) {
	jwtService := myJWT.NewJWTService(secretKey)

	jwtService.SignedStringFunc = func(token *jwt.Token, secretKey string) (string, error) {
		return "", errors.New("тестовая ошибка подписи")
	}

	tokenResp, err := jwtService.GenerateToken(auth.RoleModerator)

	assert.Error(t, err)
	assert.Nil(t, tokenResp)
	assert.EqualError(t, err, "тестовая ошибка подписи")
}

func TestValidateToken_InvalidToken(t *testing.T) {
	jwtService := myJWT.NewJWTService("wrongkey")

	goodService := myJWT.NewJWTService(secretKey)
	tokenResp, err := goodService.GenerateToken(auth.RoleEmployee)
	require.NoError(t, err)

	claims, err := jwtService.ValidateToken(tokenResp.Token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}
