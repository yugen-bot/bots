package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Game struct{ ent.Schema }

func (Game) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "Game"}}
}

func (Game) Fields() []ent.Field {
	return []ent.Field{
		field.String("guildID").
			StorageKey("guildId"),
		field.String("lastMessageID").
			StorageKey("lastMessageId").
			Optional().
			Nillable(),
		field.String("word").
			StorageKey("word"),
		field.Time("endingAt").
			StorageKey("endingAt"),
		field.Int("number").
			StorageKey("number").
			Default(1),
		field.Bool("scheduleStarted").
			StorageKey("scheduleStarted").
			Default(true),
		field.Enum("status").
			Values("IN_PROGRESS", "FAILED", "COMPLETED", "OUT_OF_TIME").
			Default("IN_PROGRESS").
			StorageKey("status"),
		field.JSON("meta", json.RawMessage{}).
			StorageKey("meta").
			Default(json.RawMessage("{}")),
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

func (Game) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("guesses", Guess.Type),
	}
}

func (Game) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guildID").
			StorageKey("Game_guildId_idx"),
	}
}
