package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

var TokenRegexp = regexp.MustCompile(`^[A-Za-z0-9-_=]{43,44}$`)

type Token struct {
	ent.Schema
}

func (Token) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (Token) Fields() []ent.Field {
	return []ent.Field{
		field.String("Token").
			Immutable().
			Match(TokenRegexp).
			Sensitive().
			Unique(),
	}
}

func (Token) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("User", User.Type).Ref("Tokens").Unique(),
	}
}

func (Token) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("Token").Unique(),
		index.Edges("User"),
	}
}
