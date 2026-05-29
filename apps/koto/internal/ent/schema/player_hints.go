package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PlayerHints struct{ ent.Schema }

func (PlayerHints) Fields() []ent.Field {
	return []ent.Field{
		field.String("userID"),
		field.Float("hints").
			Default(0),
		field.Float("maxHints").
			Default(5),
		field.Time("lastVoteTime").
			Optional().
			Nillable(),
		field.Time("createdAt").
			Default(time.Now).
			Immutable(),
		field.Time("updatedAt").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (PlayerHints) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("userID"),
	}
}
