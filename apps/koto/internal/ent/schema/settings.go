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
		field.String("channelID").
			Optional().
			Nillable(),
		field.String("pingRoleID").
			Optional().
			Nillable(),
		field.Bool("pingOnlyNew").
			Default(true),
		field.Bool("membersCanStart").
			Default(false),
		field.Int("cooldown").
			Default(600),
		field.Bool("enableBackToBackCooldown").
			Default(false),
		field.Int("backToBackCooldown").
			Default(600),
		field.Bool("informCooldownAfterGuess").
			Default(false),
		field.Int("frequency").
			Default(60),
		field.Int("timeLimit").
			Default(60),
		field.Bool("autoStart").
			Default(false),
		field.Bool("startAfterFirstGuess").
			Default(false),
		field.Float("hints").
			Default(0),
		field.Float("maxHints").
			Default(5),
		field.Float("hintsUsed").
			Default(0),
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
		index.Fields("guildID").
			Unique(),
		index.Fields("guildID", "channelID"),
	}
}
