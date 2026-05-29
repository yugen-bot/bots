package schema

import (
	"encoding/json"
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
		field.String("lastMessageID").
			Optional().
			Nillable(),
		field.String("word"),
		field.Time("endingAt"),
		field.Int("number").
			Default(1),
		field.Bool("scheduleStarted").
			Default(true),
		field.Enum("status").
			Values("IN_PROGRESS", "FAILED", "COMPLETED", "OUT_OF_TIME").
			Default("IN_PROGRESS"),
		field.JSON("meta", json.RawMessage{}).
			Default(json.RawMessage("{}")),
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
		edge.To("guesses", Guess.Type),
	}
}

func (Game) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guildID"),
	}
}
