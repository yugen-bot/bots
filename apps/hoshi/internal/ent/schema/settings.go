package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Settings struct{ ent.Schema }

func (Settings) Fields() []ent.Field {
	return []ent.Field{
		field.String("guildID"),
		field.String("botUpdatesChannelID").
			Optional().
			Nillable(),
		field.Int("treshold").
			Default(3),
		field.Bool("self").
			Default(false),
		field.JSON("ignoredChannelIds", []string{}).
			Default([]string{}),
		field.Time("createdAt").
			Default(time.Now).
			Immutable(),
		field.Time("updatedAt").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Settings) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guildID").
			Unique(),
	}
}
