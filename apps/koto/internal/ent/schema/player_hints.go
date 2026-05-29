package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PlayerHints struct{ ent.Schema }

func (PlayerHints) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "PlayerHints"}}
}

func (PlayerHints) Fields() []ent.Field {
	return []ent.Field{
		field.String("userID").
			StorageKey("userId"),
		field.Float("hints").
			StorageKey("hints").
			Default(0),
		field.Float("maxHints").
			StorageKey("maxHints").
			Default(5),
		field.Time("lastVoteTime").
			StorageKey("lastVoteTime").
			Optional().
			Nillable(),
		field.Time("createdAt").
			StorageKey("createdAt").
			Default(time.Now).
			Immutable(),
		field.Time("updatedAt").
			StorageKey("updatedAt").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (PlayerHints) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("userID").
			StorageKey("PlayerHints_userId_idx"),
	}
}
