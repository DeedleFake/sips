package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/DeedleFake/sips"
)

type Pin struct {
	ent.Schema
}

func (Pin) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (Pin) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("Status").
			GoType(sips.Queued),
		field.String("Name").
			NotEmpty(),
		field.Strings("Origins").
			Optional(),
	}
}

func (Pin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("User", User.Type).Ref("Pins").Unique(),
	}
}
