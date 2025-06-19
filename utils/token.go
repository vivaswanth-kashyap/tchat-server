package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// var JWTSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
// temp fix
var JWTSecretKey = []byte("dM7QUS3LRivJkzx9zuWXpzBVlk6u96/vbT5ZjvR3GMc=")

func init() {
	if len(JWTSecretKey) == 0 {
		fmt.Println("WARNING: JWT_SECRET_KEY environment variable not set. Using a default development key.")
		JWTSecretKey = []byte("super_secret_dev_key_dont_use_in_production")
	}

	fmt.Printf("DEBUG: JWTSecretKey length: %d. Key used (first 10 chars): %s\n", len(JWTSecretKey), string(JWTSecretKey[:min(10, len(JWTSecretKey))]))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GenerateAccessToken(userID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWTSecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}
	return tokenString, nil
}

func GenerateRefreshToken() (uuid.UUID, error) {
	refreshToken := uuid.New()
	return refreshToken, nil
}

func ParseAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		return JWTSecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid access token")
	}

	return claims, nil
}
