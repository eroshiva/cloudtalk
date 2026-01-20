// Package db implements a client for interaction with DB.
package db

import (
	"context"
	"fmt"

	"github.com/eroshiva/cloudtalk/internal/ent"
	"github.com/eroshiva/cloudtalk/internal/ent/product"
	"github.com/eroshiva/cloudtalk/internal/ent/review"
	"github.com/eroshiva/cloudtalk/pkg/logger"
	"github.com/google/uuid"
)

const (
	productPrefix = "product-"
	reviewPrefix  = "review-"
)

var zlog = logger.NewLogger("db-client")

// CreateProduct creates Product resource.
func CreateProduct(ctx context.Context, client *ent.Client, name, description, price string) (*ent.Product, error) {
	// input parameters sanity check
	if name == "" {
		err := fmt.Errorf("product name is not specified")
		zlog.Error().Err(err).Send()
		return nil, err
	}
	if description == "" {
		err := fmt.Errorf("product description is not specified")
		zlog.Error().Err(err).Send()
		return nil, err
	}
	if price == "" {
		err := fmt.Errorf("product price is not specified")
		zlog.Error().Err(err).Send()
		return nil, err
	}
	zlog.Debug().Msgf("Creating product %s", name)

	// generating random ID for the Product resource
	id := productPrefix + uuid.NewString()

	p, err := client.Product.Create().
		SetID(id).
		SetName(name).
		SetDescription(description).
		SetPrice(price).
		SetAverageRating(0). // created product doesn't have any reviews yet, setting ratings value to 0
		Save(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to create product %s", name)
		return nil, err
	}

	return p, nil
}

// GetProductByID retrieves Product resource by its ID.
func GetProductByID(ctx context.Context, client *ent.Client, id string) (*ent.Product, error) {
	zlog.Debug().Msgf("Retrieving product by ID (%s)", id)
	p, err := client.Product.
		Query().
		Where(product.ID(id)).
		// eager-loading reviews as well
		WithReviews().
		Only(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to retrieve product with ID (%s)", id)
		return nil, err
	}
	return p, nil
}

// GetProductByIDTx retrieves Product resource by its ID. Does not commit transaction!
// This function should NOT be used in the production.
func GetProductByIDTx(ctx context.Context, tx *ent.Tx, id string) (*ent.Product, error) {
	zlog.Debug().Msgf("Retrieving product by ID (%s)", id)
	p, err := tx.Product.Query().
		Where(product.ID(id)).
		// eager-loading reviews as well
		WithReviews().
		ForUpdate(). // locking
		Only(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to retrieve product with ID (%s)", id)
		return nil, rollback(tx, err)
	}
	return p, nil
}

// EditProduct updates all provided non-nil fields in Product resource.
func EditProduct(ctx context.Context, client *ent.Client, id string, name, description, price string) (*ent.Product, error) {
	zlog.Debug().Msgf("Editing product (%s)", id)

	p, err := GetProductByID(ctx, client, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		p.Name = name
	}
	if description != "" {
		p.Description = description
	}
	if price != "" {
		p.Price = price
	}
	numAfNodes, err := client.Product.Update().
		Where(product.ID(id)).
		SetName(p.Name).
		SetDescription(p.Description).
		SetPrice(p.Price).
		Save(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to edit product")
		return nil, err
	}

	if numAfNodes != 1 {
		// something bad has happened, returning error
		newErr := fmt.Errorf("update of product didn't return error, number of affected nodes is %d", numAfNodes)
		zlog.Error().Err(newErr).Send()
		return nil, err
	}
	return p, nil
}

// ListProducts retrieves all Products available in the system.
func ListProducts(ctx context.Context, client *ent.Client) ([]*ent.Product, error) {
	zlog.Debug().Msgf("Retrieving all products")

	ps, err := client.Product.Query().
		// Eager-loading edges
		WithReviews().
		All(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to retrieve all products")
		return nil, err
	}
	return ps, nil
}

// DeleteProductByID removes Product resource with provided ID from the DB.
func DeleteProductByID(ctx context.Context, client *ent.Client, id string) error {
	zlog.Debug().Msgf("Deleting product with ID (%s)", id)
	_, err := client.Product.Delete().Where(product.ID(id)).Exec(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to delete product with ID (%s)", id)
		return err
	}
	return nil
}

// rollback calls to tx.Rollback and wraps the given error
// with the rollback error if occurred.
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %v", err, rerr)
		zlog.Error().Err(err).Msgf("Failed to rollback transaction")
	}
	return err
}

// CreateReview creates a Review resource.
func CreateReview(ctx context.Context, client *ent.Client, name, lastName, text string, rating int32, productID string) (
	*ent.Review, error,
) {
	// input parameters sanity check
	if name == "" {
		err := fmt.Errorf("reviewer's name is not specified")
		zlog.Error().Err(err).Send()
		return nil, err
	}
	if lastName == "" {
		err := fmt.Errorf("reviewer's last name is not specified")
		zlog.Error().Err(err).Send()
		return nil, err
	}
	if text == "" {
		err := fmt.Errorf("text of the review is not specified")
		zlog.Error().Err(err).Send()
		return nil, err
	}
	if rating < 1 || rating > 5 {
		err := fmt.Errorf("review's rating is out of range")
		zlog.Error().Err(err).Msgf("Review's rating must be between 1 and 5, but has %d", rating)
		return nil, err
	}
	if productID == "" {
		err := fmt.Errorf("review's product is not specified")
		zlog.Error().Err(err).Send()
		return nil, err
	}

	zlog.Debug().Msgf("Creating review by %s %s for product with ID (%s)", name, lastName, productID)

	// retrieving full Product resource first
	p, err := GetProductByID(ctx, client, productID)
	if err != nil {
		zlog.Err(err).Msgf("Failed to retrieve product with ID (%s)", productID)
		return nil, err
	}
	p.Edges.Reviews = nil // no need to carry over inner references

	// generating random ID for the Review resource
	id := reviewPrefix + uuid.NewString()

	// get transaction
	tx, err := client.Tx(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to create transaction")
		return nil, err
	}

	// create review in the transaction
	r, err := tx.Review.Create().
		SetID(id).
		SetFirstName(name).
		SetLastName(lastName).
		SetReviewText(text).
		SetRating(rating).
		SetProduct(p).
		Save(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to create review by %s %s for product with ID (%s)", name, lastName, productID)
		return nil, rollback(tx, err)
	}

	// recalculate average rating during the same transaction
	_, err = updateProductAverageRating(ctx, tx, productID)
	if err != nil {
		return nil, rollback(tx, err)
	}

	// if all operations succeed, commit the transaction.
	if err = tx.Commit(); err != nil {
		zlog.Error().Err(err).Msgf("Failed to commit transaction")
		return nil, err
	}

	return r, nil
}

// GetReviewByID retrieves Review resource by its ID.
func GetReviewByID(ctx context.Context, client *ent.Client, id string) (*ent.Review, error) {
	zlog.Debug().Msgf("Retrieving review by ID (%s)", id)
	r, err := client.Review.Query().
		Where(review.ID(id)).
		WithProduct(). // eager-loading Product resource
		Only(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to retrieve review by ID (%s)", id)
		return nil, err
	}
	return r, nil
}

// GetReviewsByProductID retrieves review resource by provided Product resource ID.
func GetReviewsByProductID(ctx context.Context, client *ent.Client, id string) ([]*ent.Review, error) {
	zlog.Debug().Msgf("Retrieving all reviews for Product with ID (%s)", id)
	rs, err := client.Review.Query().
		Where(review.HasProductWith(product.ID(id))).
		All(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to retrieve all reviews for Product with ID (%s)", id)
		return nil, err
	}
	return rs, nil
}

// EditReview updates all provided non-nil fields of Review resource.
func EditReview(ctx context.Context, client *ent.Client, id string, name, lastName, text string, rating int32) (*ent.Review, error) {
	zlog.Debug().Msgf("Editing review (%s)", id)
	r, err := GetReviewByID(ctx, client, id) // Product resource is eager-loaded
	if err != nil {
		return nil, err
	}
	if name != "" {
		r.FirstName = name
	}
	if lastName != "" {
		r.LastName = lastName
	}
	if text != "" {
		r.ReviewText = text
	}
	if rating >= 1 && rating <= 5 {
		r.Rating = rating
	}
	// product is not allowed to be manipulated

	// get transaction
	tx, err := client.Tx(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to create transaction")
		return nil, err
	}

	// update review resource
	numAfNodes, err := tx.Review.Update().
		Where(review.ID(id)).
		SetFirstName(r.FirstName).
		SetLastName(r.LastName).
		SetReviewText(r.ReviewText).
		SetRating(r.Rating).
		Save(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to edit review")
		return nil, rollback(tx, err)
	}

	if numAfNodes != 1 {
		// something bad has happened, returning error
		newErr := fmt.Errorf("update of review didn't return error, number of affected nodes is %d", numAfNodes)
		zlog.Error().Err(newErr).Send()
		return nil, rollback(tx, err)
	}

	// recalculate average rating during the same transaction
	_, err = updateProductAverageRating(ctx, tx, r.Edges.Product.ID)
	if err != nil {
		return nil, err
	}

	// if all operations succeed, commit the transaction.
	if err = tx.Commit(); err != nil {
		zlog.Error().Err(err).Msgf("Failed to commit transaction")
		return nil, err
	}
	return r, nil
}

// DeleteReviewByID removes Review resource with provided ID from the DB.
func DeleteReviewByID(ctx context.Context, client *ent.Client, id, productID string) error {
	zlog.Debug().Msgf("Deleting review with ID (%s)", id)
	// get transaction
	tx, err := client.Tx(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to create transaction")
		return err
	}

	// delete of Review resource
	_, err = tx.Review.Delete().Where(review.ID(id)).Exec(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to delete review with ID (%s)", id)
		return rollback(tx, err)
	}

	// recalculate average rating during the same transaction
	_, err = updateProductAverageRating(ctx, tx, productID)
	if err != nil {
		return err
	}

	// if all operations succeed, commit the transaction.
	if err = tx.Commit(); err != nil {
		zlog.Error().Err(err).Msgf("Failed to commit transaction")
		return err
	}
	return nil
}

// updateProductAverageRating performs recalculation of average rating during the same transaction.
func updateProductAverageRating(ctx context.Context, tx *ent.Tx, productID string) (*ent.Product, error) {
	zlog.Info().Msgf("Updating average product rating for product (%s)", productID)
	// fetch the Product resource by ID during provided transaction and perform a database lock.
	p, err := tx.Product.Query().
		Where(product.ID(productID)).
		WithReviews(). // eager-loading all reviews
		ForUpdate().   // pessimistic locking => lock the product row for the duration of this transaction
		Only(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to retrieve product with ID (%s)", productID)
		return nil, rollback(tx, err)
	}

	// calculate the sum of ratings and the total count.
	var totalRating int32
	for _, r := range p.Edges.Reviews {
		totalRating += r.Rating
	}

	reviewCount := len(p.Edges.Reviews) // total number of reviews
	// computing average rating
	newAverage := 0.0
	if reviewCount > 0 {
		newAverage = float64(totalRating) / float64(reviewCount)
	}

	// updating Product resource with the newly calculated values
	updatedProduct, err := tx.Product.UpdateOne(p).
		SetAverageRating(newAverage).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update average rating for product with ID (%s)", productID)
		return nil, rollback(tx, err)
	}

	return updatedProduct, nil
}
