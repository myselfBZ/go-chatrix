package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthMock struct {
}

const secret = "test"

var testClaims = jwt.MapClaims{
	"aud": "test-aud",
	"iss": "test-aud",
	"sub": int64(1),
	"exp": time.Now().Add(time.Hour).Unix(),
}

func (a *AuthMock) ValidateToken(token string) (*jwt.Token, error) {
    return nil, nil
}

func (a *AuthMock) GenerateToken(claims jwt.Claims) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims)

    tokenString, _ := token.SignedString([]byte(secret))
    return tokenString, nil
}
