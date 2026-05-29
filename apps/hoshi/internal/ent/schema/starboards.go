package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Starboards struct{ ent.Schema }

func (Starboards) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "Starboards"}}
}

func (Starboards) Fields() []ent.Field {
	return []ent.Field{
		field.String("guildID").
			StorageKey("guildId"),
		field.String("sourceEmoji").
			StorageKey("sourceEmoji").
			Default("⭐"),
		field.String("sourceChannelID").
			StorageKey("sourceChannelId").
			Optional().
			Nillable(),
		field.String("targetChannelID").
			StorageKey("targetChannelId"),
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
