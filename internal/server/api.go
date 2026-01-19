package server

import (
	"context"
	"fmt"

	apiv1 "github.com/eroshiva/cloudtalk/api/v1"
	"github.com/eroshiva/cloudtalk/pkg/client/db"
)

// CreateProduct creates a Product resource in the DB.
func (srv *server) CreateProduct(ctx context.Context, req *apiv1.CreateProductRequest) (*apiv1.CreateProductResponse, error) {
	zlog.Info().Msgf("Creating product %s", req.GetProduct().GetName())
	// sanity check
	if req.GetProduct() == nil {
		err := fmt.Errorf("product resource is not specified")
		zlog.Error().Err(err).Msg("Failed to create product")
		return nil, err
	}

	// creating product
	p, err := db.CreateProduct(ctx, srv.dbClient, req.GetProduct().GetName(),
		req.GetProduct().GetDescription(), req.GetProduct().GetPrice())
	if err != nil {
		return nil, err
	}

	return &apiv1.CreateProductResponse{
		Product: &apiv1.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		},
	}, nil
}

// GetProductByID retrieves Product resource from the DB by specified ID.
func (srv *server) GetProductByID(ctx context.Context, req *apiv1.GetProductByIDRequest) (*apiv1.GetProductByIDResponse, error) {
	zlog.Info().Msgf("Retrieving product by its ID (%s)", req.GetId())
	// sanity check
	if req.GetId() == "" {
		err := fmt.Errorf("ID is not specified")
		zlog.Error().Err(err).Msg("Failed to retrieve product by its ID")
		return nil, err
	}

	// retrieving product by ID
	p, err := db.GetProductByID(ctx, srv.dbClient, req.GetId())
	if err != nil {
		return nil, err
	}

	return &apiv1.GetProductByIDResponse{
		Product: ConvertProductResourceToProtobuf(p),
	}, nil
}

// EditProduct updates specified fields in Product resource in the DB.
func (srv *server) EditProduct(ctx context.Context, req *apiv1.EditProductRequest) (*apiv1.EditProductResponse, error) {
	zlog.Info().Msgf("Editing product (%s)", req.GetProduct().GetId())
	// sanity check
	if req.GetProduct() == nil {
		err := fmt.Errorf("product resource is not specified")
		zlog.Error().Err(err).Msg("Failed to edit product")
		return nil, err
	}
	if req.GetProduct().GetId() == "" {
		err := fmt.Errorf("product ID is not specified")
		zlog.Error().Err(err).Msg("Failed to edit product")
		return nil, err
	}

	// updating product in DB
	updP, err := db.EditProduct(ctx, srv.dbClient, req.GetProduct().GetId(), req.GetProduct().GetName(),
		req.GetProduct().GetDescription(), req.GetProduct().GetPrice())
	if err != nil {
		return nil, err
	}

	return &apiv1.EditProductResponse{
		Product: ConvertProductResourceToProtobuf(updP),
	}, nil
}

// DeleteProduct deletes Product resource from the DB.
func (srv *server) DeleteProduct(ctx context.Context, req *apiv1.DeleteProductRequest) error {
	zlog.Info().Msgf("Deleting product (%s)", req.GetId())
	// sanity check
	if req.GetId() == "" {
		err := fmt.Errorf("product ID is not specified")
		zlog.Error().Err(err).Msg("Failed to delete product")
		return err
	}

	// deleting product from DB
	err := db.DeleteProductByID(ctx, srv.dbClient, req.GetId())
	if err != nil {
		return err
	}
	return nil
}

// ListProducts lists all Product resources available in the DB.
func (srv *server) ListProducts(ctx context.Context) ([]*apiv1.Product, error) {
	zlog.Info().Msgf("Listing all products")
	ps, err := db.ListProducts(ctx, srv.dbClient)
	if err != nil {
		return nil, err
	}

	resp := make([]*apiv1.Product, 0)
	for _, p := range ps {
		resp = append(resp, ConvertProductResourceToProtobuf(p))
	}
	return resp, nil
}
