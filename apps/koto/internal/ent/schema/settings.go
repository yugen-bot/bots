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
		field.String("pingRoleID").
			StorageKey("pingRoleId").
			Optional().
			Nillable(),
		field.Bool("pingOnlyNew").
			StorageKey("pingOnlyNew").
			Default(true),
		field.Bool("membersCanStart").
			StorageKey("membersCanStart").
			Default(false),
		field.Int("cooldown").
			StorageKey("cooldown").
			Default(600),
		field.Bool("enableBackToBackCooldown").
			StorageKey("enableBackToBackCooldown").
			Default(false),
		field.Int("backToBackCooldown").
			StorageKey("backToBackCooldown").
			Default(600),
		field.Bool("informCooldownAfterGuess").
			StorageKey("informCooldownAfterGuess").
			Default(false),
		field.Int("frequency").
			StorageKey("frequency").
			Default(60),
		field.Int("timeLimit").
			StorageKey("timeLimit").
			Default(60),
		field.Bool("autoStart").
			StorageKey("autoStart").
			Default(false),
		field.Bool("startAfterFirstGuess").
			StorageKey("startAfterFirstGuess").
			Default(false),
		field.Float("hints").
			StorageKey("hints").
			Default(0),
		field.Float("maxHints").
			StorageKey("maxHints").
			Default(5),
		field.Float("hintsUsed").
			StorageKey("hintsUsed").
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

func (Settings) Edges() []ent.Edge {
	return nil
}

func (Settings) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guildID").
			Unique().
			StorageKey("Settings_guildId_key"),
		index.Fields("guildID", "channelID").
			StorageKey("Settings_guildId_channelId_idx"),
	}
}
