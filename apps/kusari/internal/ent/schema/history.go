package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type History struct{ ent.Schema }

func (History) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "History"}}
}

func (History) Fields() []ent.Field {
	return []ent.Field{
		field.String("userID").
			StorageKey("userId"),
		field.String("messageID").
			StorageKey("messageId").
			Optional().
			Nillable(),
		field.Int("gameID").
			StorageKey("gameId"),
		field.String("word").
			StorageKey("word"),
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
		index.Fields("userID", "gameID").
			StorageKey("History_userId_gameId_idx"),
	}
}
