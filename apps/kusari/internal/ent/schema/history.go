package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type History struct{ ent.Schema }

func (History) Fields() []ent.Field {
	return []ent.Field{
		field.String("userID"),
		field.String("messageID").
			Optional().
			Nillable(),
		field.Int("gameID"),
		field.String("word"),
		field.Time("createdAt").
			Default(time.Now).
			Immutable(),
		field.Time("updatedAt").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (History) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("game", Game.Type).
			Ref("history").
			Field("gameID").
			Required().
			Unique(),
	}
}

func (History) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("userID", "gameID"),
	}
}
