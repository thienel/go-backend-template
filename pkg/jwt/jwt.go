package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/thienel/go-backend-template/pkg/config"
)

// Claims represents the JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var (
	ErrTokenInvalid = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

// GenerateAccessToken generates a new access token
func GenerateAccessToken(userID uint, username, role string) (string, error) {
	cfg := config.GetConfig()
	expiry := time.Now().Add(time.Duration(cfg.JWT.AccessExpiryMinutes) * time.Minute)

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// GenerateRefreshToken generates a new refresh token
func GenerateRefreshToken(userID uint, username, role string) (string, error) {
	cfg := config.GetConfig()
	expiry := time.Now().Add(time.Duration(cfg.JWT.RefreshExpiryHours) * time.Hour)

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ValidateToken validates a JWT token and returns its claims
func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.GetConfig()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// GetAccessExpiryTime returns the access token expiry time
func GetAccessExpiryTime() time.Time {
	cfg := config.GetConfig()
	return time.Now().Add(time.Duration(cfg.JWT.AccessExpiryMinutes) * time.Minute)
}

// GetRefreshExpiryTime returns the refresh token expiry time
func GetRefreshExpiryTime() time.Time {
	cfg := config.GetConfig()
	return time.Now().Add(time.Duration(cfg.JWT.RefreshExpiryHours) * time.Hour)
}
