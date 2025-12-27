package service

import "github.com/thienel/go-backend-template/internal/domain/valueobject"

// JWTService defines JWT operations
type JWTService interface {
	GenerateAccessToken(userID uint, username, role string) (string, error)
	GenerateRefreshToken(userID uint, username, role string) (string, error)
	ValidateToken(tokenString string) (*valueobject.JWTClaims, error)
	GetAccessExpirySeconds() int
	GetRefreshExpirySeconds() int
}
