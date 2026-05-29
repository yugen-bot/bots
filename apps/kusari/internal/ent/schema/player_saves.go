package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PlayerSaves struct{ ent.Schema }

func (PlayerSaves) Fields() []ent.Field {
	return []ent.Field{
		field.String("userID"),
		field.Float("saves").
			Default(0),
		field.Float("maxSaves").
			Default(2),
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

func (PlayerSaves) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("userID"),
	}
}
