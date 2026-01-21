package server

import (
	"context"
	"fmt"

	apiv1 "github.com/eroshiva/cloudtalk/api/v1"
	"github.com/eroshiva/cloudtalk/pkg/client/db"
	"github.com/eroshiva/cloudtalk/pkg/rabbitmq"
	"google.golang.org/protobuf/types/known/emptypb"
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

	// updating cache
	srv.cache.SetProduct(p)

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

	// checking cache first
	if p, ok := srv.cache.GetProduct(req.GetId()); ok {
		zlog.Info().Msgf("Product %s found in cache", req.GetId())
		return &apiv1.GetProductByIDResponse{
			Product: ConvertProductResourceToProtobuf(p),
		}, nil
	}

	// retrieving product by ID
	p, err := db.GetProductByID(ctx, srv.dbClient, req.GetId())
	if err != nil {
		return nil, err
	}

	// setting cache
	srv.cache.SetProduct(p)

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

	// invalidating cache
	srv.cache.DeleteProduct(req.GetProduct().GetId())

	return &apiv1.EditProductResponse{
		Product: ConvertProductResourceToProtobuf(updP),
	}, nil
}

// DeleteProduct deletes Product resource from the DB.
func (srv *server) DeleteProduct(ctx context.Context, req *apiv1.DeleteProductRequest) (*emptypb.Empty, error) {
	zlog.Info().Msgf("Deleting product (%s)", req.GetId())
	// sanity check
	if req.GetId() == "" {
		err := fmt.Errorf("product ID is not specified")
		zlog.Error().Err(err).Msg("Failed to delete product")
		return nil, err
	}

	// deleting product from DB
	err := db.DeleteProductByID(ctx, srv.dbClient, req.GetId())
	if err != nil {
		return nil, err
	}

	// invalidating cache
	srv.cache.DeleteProduct(req.GetId())

	return &emptypb.Empty{}, nil
}

// ListProducts lists all Product resources available in the DB.
func (srv *server) ListProducts(ctx context.Context, _ *emptypb.Empty) (*apiv1.ListProductsResponse, error) {
	zlog.Info().Msgf("Listing all products")
	ps, err := db.ListProducts(ctx, srv.dbClient)
	if err != nil {
		return nil, err
	}

	resp := make([]*apiv1.Product, 0)
	for _, p := range ps {
		resp = append(resp, ConvertProductResourceToProtobuf(p))
	}
	return &apiv1.ListProductsResponse{
		Products: resp,
	}, nil
}

// CreateReview creates Review resource in the DB. Enough to specify only Product ID in the Product field.
func (srv *server) CreateReview(ctx context.Context, req *apiv1.CreateReviewRequest) (*apiv1.CreateReviewResponse, error) {
	zlog.Info().Msgf("Creating review %s", req.GetReview().GetId())
	// sanity check
	if req.GetReview() == nil {
		err := fmt.Errorf("review resource is not specified")
		zlog.Error().Err(err).Msg("Failed to create review")
		return nil, err
	}
	if req.GetReview().GetFirstName() == "" || req.GetReview().GetLastName() == "" {
		err := fmt.Errorf("reviewer's identity is not specified")
		zlog.Error().Err(err).Msg("Failed to create review")
		return nil, err
	}
	if req.GetReview().GetRating() < 1 || req.GetReview().GetRating() > 5 {
		err := fmt.Errorf("reviewer's rating is out of range")
		zlog.Error().Err(err).Msg("Failed to create review")
		return nil, err
	}
	// allowing to create review with empty text

	// creating review
	r, err := db.CreateReview(ctx, srv.dbClient, req.GetReview().GetFirstName(), req.GetReview().GetLastName(),
		req.GetReview().GetReviewText(), req.GetReview().GetRating(), req.GetReview().GetProduct().GetId())
	if err != nil {
		return nil, err
	}

	// invalidating cache
	srv.cache.DeleteReviews(req.GetReview().GetProduct().GetId())
	srv.cache.DeleteProduct(req.GetReview().GetProduct().GetId()) // removing product entry so fresh data can be fetched during the Get operation

	// publishing event that review was created
	err = rabbitmq.PublishMessage(ctx, srv.rabbitMQChannel, ComposeEventOnReviewChange("created",
		r.Rating, r.FirstName, r.LastName, req.GetReview().GetProduct().GetId()))
	if err != nil {
		return nil, err
	}

	return &apiv1.CreateReviewResponse{
		Review: ConvertReviewResourceToProtobuf(r),
	}, nil
}

// GetReviewsByProductID retrieves Review resource by Product ID from DB.
func (srv *server) GetReviewsByProductID(ctx context.Context, req *apiv1.GetReviewsByProductIDRequest) (*apiv1.GetReviewsByProductIDResponse, error) {
	zlog.Info().Msgf("Retrieving Review by Product ID (%s)", req.GetId())
	// sanity check
	if req.GetId() == "" {
		err := fmt.Errorf("product ID is not specified")
		zlog.Error().Err(err).Msg("Failed to retrieve review by product ID")
		return nil, err
	}

	// checking cache first
	if rs, ok := srv.cache.GetReviews(req.GetId()); ok {
		zlog.Info().Msgf("Reviews for product %s found in cache", req.GetId())
		reviews := make([]*apiv1.Review, 0)
		for _, r := range rs {
			reviews = append(reviews, ConvertReviewResourceToProtobuf(r))
		}
		return &apiv1.GetReviewsByProductIDResponse{
			Reviews: reviews,
		}, nil
	}

	// retrieving resource
	rs, err := db.GetReviewsByProductID(ctx, srv.dbClient, req.GetId())
	if err != nil {
		return nil, err
	}

	// setting/updating cache
	srv.cache.SetReviews(req.GetId(), rs)

	// converting back to Protobuf format
	reviews := make([]*apiv1.Review, 0)
	for _, r := range rs {
		reviews = append(reviews, ConvertReviewResourceToProtobuf(r))
	}
	return &apiv1.GetReviewsByProductIDResponse{
		Reviews: reviews,
	}, nil
}

// EditReview updates specified fields of the Review resource in the DB.
func (srv *server) EditReview(ctx context.Context, req *apiv1.EditReviewRequest) (*apiv1.EditReviewResponse, error) {
	zlog.Info().Msgf("Editing review (%s)", req.GetReview().GetId())
	// sanity check
	if req.GetReview() == nil {
		err := fmt.Errorf("review resource is not specified")
		zlog.Error().Err(err).Msg("Failed to edit review")
		return nil, err
	}
	if req.GetReview().GetId() == "" {
		err := fmt.Errorf("review ID is not specified")
		zlog.Error().Err(err).Msg("Failed to edit review")
		return nil, err
	}

	// updating review resource
	updR, err := db.EditReview(ctx, srv.dbClient, req.GetReview().GetId(), req.GetReview().GetFirstName(),
		req.GetReview().GetLastName(), req.GetReview().GetReviewText(), req.GetReview().GetRating())
	if err != nil {
		return nil, err
	}

	// invalidating cache
	srv.cache.DeleteReviews(updR.Edges.Product.ID)
	srv.cache.DeleteProduct(updR.Edges.Product.ID) // removing product entry so fresh data can be fetched during the Get operation

	// publishing event that review was modified
	err = rabbitmq.PublishMessage(ctx, srv.rabbitMQChannel, ComposeEventOnReviewChange("modified", updR.Rating,
		updR.FirstName, updR.LastName, updR.Edges.Product.ID))
	if err != nil {
		return nil, err
	}

	return &apiv1.EditReviewResponse{
		Review: ConvertReviewResourceToProtobuf(updR),
	}, nil
}

// DeleteReview removes specified Review resource from the DB.
func (srv *server) DeleteReview(ctx context.Context, req *apiv1.DeleteReviewRequest) (*emptypb.Empty, error) {
	zlog.Info().Msgf("Deleting review (%s)", req.GetId())
	// sanity check
	if req.GetId() == "" {
		err := fmt.Errorf("review ID is not specified")
		zlog.Error().Err(err).Msg("Failed to delete review")
		return nil, err
	}

	// Querying review first to get a fancy-published message - in favor of unified published messages structure.
	// Normally, this should be brought to forum with colleagues and defined what precisely we want to publish on bus. My take - IDs are simple enough.
	// For the sake of better visibility for this task leaving it this way.
	r, err := db.GetReviewByID(ctx, srv.dbClient, req.GetId())
	if err != nil {
		return nil, err
	}

	// removing review resource
	err = db.DeleteReviewByID(ctx, srv.dbClient, req.GetId(), r.Edges.Product.ID)
	if err != nil {
		return nil, err
	}

	// invalidating cache
	srv.cache.DeleteReviews(r.Edges.Product.ID)
	srv.cache.DeleteProduct(r.Edges.Product.ID) // removing product entry so fresh data can be fetched during the Get operation

	// publishing event that review was deleted
	err = rabbitmq.PublishMessage(ctx, srv.rabbitMQChannel, ComposeEventOnReviewChange("deleted", r.Rating, r.FirstName, r.LastName, r.Edges.Product.ID))
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
