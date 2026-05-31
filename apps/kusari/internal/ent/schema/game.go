package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Game struct{ ent.Schema }

func (Game) Fields() []ent.Field {
	return []ent.Field{
		field.String("guildID"),
		field.Enum("status").
			Values("IN_PROGRESS", "FAILED", "COMPLETED").
			Default("IN_PROGRESS"),
		field.Enum("type").
			Values("NORMAL").
			Default("NORMAL"),
		field.Bool("isHighscored").
			Default(false),
		field.Time("createdAt").
			Default(time.Now).
			Immutable(),
		field.Time("updatedAt").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Game) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("history", History.Type),
	}
}

func (Game) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guildID"),
	}
}
