// File updated by protoc-gen-ent.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Product struct {
	ent.Schema
}

func (Product) Fields() []ent.Field {
	return []ent.Field{field.String("id"), field.String("name"), field.String("description"), field.String("price"), field.String("average_rating")}
}
func (Product) Edges() []ent.Edge {
	return []ent.Edge{edge.To("reviews", Review.Type)}
}
func (Product) Annotations() []schema.Annotation {
	return nil
}
