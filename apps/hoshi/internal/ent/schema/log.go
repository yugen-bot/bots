package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type StarboardLog struct{ ent.Schema }

func (StarboardLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("guildID"),
		field.String("channelID"),
		field.String("messageID"),
		field.String("originalMessageID"),
		field.Time("createdAt").
			Default(time.Now).
			Immutable(),
		field.Time("updatedAt").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (StarboardLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("messageID").
			Unique(),
		index.Fields("originalMessageID").
			Unique(),
	}
}
