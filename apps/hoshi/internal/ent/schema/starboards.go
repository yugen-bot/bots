package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Starboards struct{ ent.Schema }

func (Starboards) Fields() []ent.Field {
	return []ent.Field{
		field.String("guildID"),
		field.String("sourceEmoji").
			Default("⭐"),
		field.String("sourceChannelID").
			Optional().
			Nillable(),
		field.String("targetChannelID"),
		field.Time("createdAt").
			Default(time.Now).
			Immutable(),
		field.Time("updatedAt").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}
