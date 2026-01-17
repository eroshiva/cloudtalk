// File updated by protoc-gen-ent.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Review struct {
	ent.Schema
}

func (Review) Fields() []ent.Field {
	return []ent.Field{field.String("id"), field.String("first_name"), field.String("last_name"), field.String("review_text"), field.Int32("rating")}
}
func (Review) Edges() []ent.Edge {
	return []ent.Edge{edge.From("product", Product.Type).Ref("reviews").Unique()}
}
func (Review) Annotations() []schema.Annotation {
	return nil
}
