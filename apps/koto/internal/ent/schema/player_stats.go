package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PlayerStats struct{ ent.Schema }

func (PlayerStats) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "PlayerStats"}}
}

func (PlayerStats) Fields() []ent.Field {
	return []ent.Field{
		field.String("userID").
			StorageKey("userId"),
		field.String("guildID").
			StorageKey("guildId"),
		field.Bool("inGuild").
			StorageKey("inGuild").
			Default(true),
		field.Int("points").
			StorageKey("points").
			Default(0),
		field.Int("participated").
			StorageKey("participated").
			Default(0),
		field.Int("wins").
			StorageKey("wins").
			Default(0),
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

func (PlayerStats) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("userID", "guildID").
			StorageKey("PlayerStats_userId_guildId_idx"),
	}
}
