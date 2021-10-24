package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("Name").
			NotEmpty().
			Unique(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("Tokens", Token.Type),
		edge.To("Pins", Pin.Type),
	}
}
