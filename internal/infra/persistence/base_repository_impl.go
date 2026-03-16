package persistence

import (
	"github.com/thienel/go-backend-template/internal/ent"
)

// BaseRepositoryImpl holds the shared ent.Client
type BaseRepositoryImpl struct {
	Client        *ent.Client
	AllowedFields map[string]bool
	EntityName    string
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(client *ent.Client, allowedFields map[string]bool, entityName string) *BaseRepositoryImpl {
	return &BaseRepositoryImpl{
		Client:        client,
		AllowedFields: allowedFields,
		EntityName:    entityName,
	}
}
