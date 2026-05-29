package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PlayerSaves struct{ ent.Schema }

func (PlayerSaves) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "PlayerSaves"}}
}

func (PlayerSaves) Fields() []ent.Field {
	return []ent.Field{
		field.String("userID").
			StorageKey("userId"),
		field.Float("saves").
			StorageKey("saves").
			Default(0),
		field.Float("maxSaves").
			StorageKey("maxSaves").
			Default(2),
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

func (PlayerSaves) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("userID").
			StorageKey("PlayerSaves_userId_idx"),
	}
}
