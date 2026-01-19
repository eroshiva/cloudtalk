package server

import (
	apiv1 "github.com/eroshiva/cloudtalk/api/v1"
	"github.com/eroshiva/cloudtalk/internal/ent"
)

func ConvertReviewResourceToProtobuf(r *ent.Review) *apiv1.Review {
	return &apiv1.Review{
		Id:         r.ID,
		FirstName:  r.FirstName,
		LastName:   r.LastName,
		ReviewText: r.ReviewText,
		Rating:     r.Rating,
	}
}

func ConvertProductResourceToProtobuf(p *ent.Product) *apiv1.Product {
	product := &apiv1.Product{
		Id:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price,
		AverageRating: p.AverageRating,
		Reviews:       make([]*apiv1.Review, 0),
	}
	// convert all reviews to Protobuf notation
	if len(p.Edges.Reviews) > 0 {
		for _, r := range p.Edges.Reviews {
			product.Reviews = append(product.Reviews, ConvertReviewResourceToProtobuf(r))
		}
	} else {
		zlog.Debug().Msgf("Product (%s) does NOT contain any reviews", p.ID)
	}
	return product
}
