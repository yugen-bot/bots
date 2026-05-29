package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
		field.String("channelID").
			StorageKey("channelId").
			Optional().
			Nillable(),
		field.Int("cooldown").
			StorageKey("cooldown").
			Default(0),
		field.Bool("math").
			StorageKey("math").
			Default(true),
		field.String("shameRoleID").
			StorageKey("shameRoleId").
			Optional().
			Nillable(),
		field.Bool("removeShameRoleAfterHighscore").
			StorageKey("removeShameRoleAfterHighscore").
			Default(false),
		field.String("lastShameUserID").
			StorageKey("lastShameUserId").
			Optional().
			Nillable(),
		field.Int("highscore").
			StorageKey("highscore").
			Default(0),
		field.Time("highscoreDate").
			StorageKey("highscoreDate").
			Optional().
			Nillable(),
		field.Float("saves").
			StorageKey("saves").
			Default(0),
		field.Float("maxSaves").
			StorageKey("maxSaves").
			Default(2),
		field.Float("savesUsed").
			StorageKey("savesUsed").
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

func (Settings) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guildID").
			Unique().
			StorageKey("Settings_guildId_key"),
		index.Fields("guildID").
			StorageKey("Settings_guildId_idx"),
	}
}
