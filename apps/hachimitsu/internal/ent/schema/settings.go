package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Settings holds the per-guild configuration for hachimitsu.
type Settings struct{ ent.Schema }

func (Settings) Fields() []ent.Field {
	return []ent.Field{
		field.String("guildID"),
		field.String("logChannelID").
			Optional().
			Nillable(),
		field.String("logPingRoleID").
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

func (Settings) Edges() []ent.Edge {
	return nil
}

func (Settings) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guildID").Unique(),
	}
}
