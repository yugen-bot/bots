package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/lib/pq"
)

type Settings struct{ ent.Schema }

func (Settings) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "Settings"}}
}

func (Settings) Fields() []ent.Field {
	return []ent.Field{
		field.String("guildID").
			StorageKey("guildId"),
		field.String("botUpdatesChannelID").
			StorageKey("botUpdatesChannelId").
			Optional().
			Nillable(),
		field.Int("treshold").
			StorageKey("treshold").
			Default(3),
		field.Bool("self").
			StorageKey("self").
			Default(false),
		field.Other("ignoredChannelIds", pq.StringArray{}).
			StorageKey("ignoredChannelIds").
			SchemaType(map[string]string{"postgres": "text[]"}).
			Default(pq.StringArray{}),
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

func (Settings) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guildID").
			Unique().
			StorageKey("Settings_guildId_key"),
	}
}
