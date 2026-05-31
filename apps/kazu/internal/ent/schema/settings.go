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
		field.Int("cooldown").
			Default(0),
		field.Bool("math").
			Default(true),
		field.String("shameRoleID").
			Optional().
			Nillable(),
		field.Bool("removeShameRoleAfterHighscore").
			Default(false),
		field.String("lastShameUserID").
			Optional().
			Nillable(),
		field.Int("highscore").
			Default(0),
		field.Time("highscoreDate").
			Optional().
			Nillable(),
		field.Float("saves").
			Default(0),
		field.Float("maxSaves").
			Default(2),
		field.Float("savesUsed").
			Default(0),
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
		index.Fields("guildID"),
	}
}
