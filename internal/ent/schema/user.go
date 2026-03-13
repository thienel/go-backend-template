package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id"),
		field.String("username").Unique().MaxLen(50).NotEmpty(),
		field.String("email").Unique().MaxLen(255).NotEmpty(),
		field.String("password").MaxLen(255).NotEmpty().Sensitive(),
		field.String("role").Default("USER").MaxLen(20),
		field.String("status").Default("ACTIVE").MaxLen(20),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
