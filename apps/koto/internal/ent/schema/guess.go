package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Guess struct{ ent.Schema }

func (Guess) Fields() []ent.Field {
	return []ent.Field{
		field.String("userID"),
		field.Int("gameID"),
		field.String("word"),
		field.Int("points").
			Default(0),
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
		index.Fields("gameID"),
	}
}
