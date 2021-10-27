package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/DeedleFake/sips"
)

var CIDRegexp = regexp.MustCompile(`^[A-Za-z0-9-_=]+$`)

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
			Default(string(sips.Queued)).
			GoType(sips.Queued),
		field.String("Name").
			NotEmpty(),
		field.String("CID").
			Match(CIDRegexp),
		field.Strings("Origins").
			Optional(),
	}
}

func (Pin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("User", User.Type).Ref("Pins").Unique(),
	}
}
