package auth

import (
	"fmt"
	"recruitment-system/config"
	"recruitment-system/models"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	UserType string `json:"user_type"`
	jwt.RegisteredClaims
}

func GenerateJWT(user *models.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(config.AppConfig.JwtExpirationHours) * time.Hour)
	claims := &Claims{
		UserID:   user.ID,
		UserType: user.UserType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JwtSecret))
}

func ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
