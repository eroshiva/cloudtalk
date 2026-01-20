package server

import (
	"fmt"
	"strings"

	apiv1 "github.com/eroshiva/cloudtalk/api/v1"
)

// CreateProductRequest is a wrapper for CreateProductRequest struct.
func CreateProductRequest(name, description, price string) *apiv1.CreateProductRequest {
	return &apiv1.CreateProductRequest{
		Product: &apiv1.Product{
			Name:        name,
			Description: description,
			Price:       price,
		},
	}
}

// GetProductByIDRequest is a wrapper for GetProductByIDRequest struct.
func GetProductByIDRequest(id string) *apiv1.GetProductByIDRequest {
	return &apiv1.GetProductByIDRequest{
		Id: id,
	}
}

// EditProductRequest is a wrapper for EditProductRequest struct.
func EditProductRequest(id, name, description, price string) *apiv1.EditProductRequest {
	return &apiv1.EditProductRequest{
		Product: &apiv1.Product{
			Id:          id,
			Name:        name,
			Description: description,
			Price:       price,
		},
	}
}

// DeleteProductRequest is a wrapper for DeleteProductRequest struct.
func DeleteProductRequest(id string) *apiv1.DeleteProductRequest {
	return &apiv1.DeleteProductRequest{
		Id: id,
	}
}

// CreateReviewRequest is a wrapper for CreateReviewRequest struct.
func CreateReviewRequest(name, lastName, reviewText string, rating int32, productID string) *apiv1.CreateReviewRequest {
	return &apiv1.CreateReviewRequest{
		Review: &apiv1.Review{
			FirstName:  name,
			LastName:   lastName,
			ReviewText: reviewText,
			Rating:     rating,
			Product: &apiv1.Product{
				Id: productID,
			},
		},
	}
}

// EditReviewRequest is a wrapper for EditReviewRequest struct.
func EditReviewRequest(id, name, lastName, reviewText string, rating int32) *apiv1.EditReviewRequest {
	return &apiv1.EditReviewRequest{
		Review: &apiv1.Review{
			Id:         id,
			FirstName:  name,
			LastName:   lastName,
			ReviewText: reviewText,
			Rating:     rating,
		},
	}
}

// DeleteReviewRequest is a wrapper for DeleteReviewRequest struct.
func DeleteReviewRequest(id string) *apiv1.DeleteReviewRequest {
	return &apiv1.DeleteReviewRequest{
		Id: id,
	}
}

// GetReviewsByProductIDRequest is a wrapper for GetReviewsByProductIDRequest struct.
func GetReviewsByProductIDRequest(id string) *apiv1.GetReviewsByProductIDRequest {
	return &apiv1.GetReviewsByProductIDRequest{
		Id: id,
	}
}

// ComposeEventOnReviewChange function composes a one-liner that is published to the RabbitMQ on any review event (addition, change, deletion).
func ComposeEventOnReviewChange(action string, rating int32, name, lastName, productID string) string {
	return fmt.Sprintf("%s: Review scoring %d from %s %s for product %s",
		strings.ToUpper(action), rating, name, lastName, productID)
}
