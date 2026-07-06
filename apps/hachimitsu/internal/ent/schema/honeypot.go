package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Honeypot holds per-channel honeypot configuration.
type Honeypot struct{ ent.Schema }

func (Honeypot) Fields() []ent.Field {
	return []ent.Field{
		field.String("guildID"),
		field.String("channelID"),
		field.JSON("ignoredRoleIDs", []string{}).
			Default([]string{}),
		field.Int("deleteMessageDays").
			Default(7),
		field.Time("createdAt").
			Default(time.Now).
			Immutable(),
		field.Time("updatedAt").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Honeypot) Edges() []ent.Edge {
	return nil
}

func (Honeypot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guildID", "channelID").Unique(),
		index.Fields("guildID"),
	}
}
