package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type StarboardLog struct{ ent.Schema }

func (StarboardLog) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "Log"}}
}

func (StarboardLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("guildID").
			StorageKey("guildId"),
		field.String("channelID").
			StorageKey("channelId"),
		field.String("messageID").
			StorageKey("messageId"),
		field.String("originalMessageID").
			StorageKey("originalMessageId"),
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

func (StarboardLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("messageID").
			Unique().
			StorageKey("Log_messageId_key"),
		index.Fields("originalMessageID").
			Unique().
			StorageKey("Log_originalMessageId_key"),
	}
}
