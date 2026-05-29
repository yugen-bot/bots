package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PlayerStats struct{ ent.Schema }

func (PlayerStats) Fields() []ent.Field {
	return []ent.Field{
		field.String("userID"),
		field.String("guildID"),
		field.Bool("inGuild").
			Default(true),
		field.Int("points").
			Default(0),
		field.Int("participated").
			Default(0),
		field.Int("wins").
			Default(0),
		field.Time("createdAt").
			Default(time.Now).
			Immutable(),
		field.Time("updatedAt").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (PlayerStats) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("userID", "guildID"),
	}
}
