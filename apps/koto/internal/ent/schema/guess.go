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

type Guess struct{ ent.Schema }

func (Guess) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "Guess"}}
}

func (Guess) Fields() []ent.Field {
	return []ent.Field{
		field.String("userID").
			StorageKey("userId"),
		field.Int("gameID").
			StorageKey("gameId"),
		field.String("word").
			StorageKey("word"),
		field.Int("points").
			StorageKey("points").
			Default(0),
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

func (Guess) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("game", Game.Type).
			Ref("guesses").
			Field("gameID").
			Required().
			Unique(),
	}
}

func (Guess) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("gameID").
			StorageKey("Guess_gameId_idx"),
	}
}
