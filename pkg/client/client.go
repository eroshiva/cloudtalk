package client

import (
	"context"
	"fmt"

	"github.com/eroshiva/cloudtalk/internal/ent"
	"github.com/eroshiva/cloudtalk/internal/ent/product"
	"github.com/eroshiva/cloudtalk/pkg/logger"
	"github.com/google/uuid"
)

const (
	productPrefix = "product-"
	reviewPrefix  = "review-"
)

var zlog = logger.NewLogger("db-client")

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
		Save(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to create product %s", name)
		return nil, err
	}

	return p, nil
}

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

func DeleteProductByID(ctx context.Context, client *ent.Client, id string) error {
	zlog.Debug().Msgf("Deleting product with ID (%s)", id)
	_, err := client.Product.Delete().Where(product.ID(id)).Exec(ctx)
	if err != nil {
		zlog.Err(err).Msgf("Failed to delete product with ID (%s)", id)
		return err
	}
	return nil
}
