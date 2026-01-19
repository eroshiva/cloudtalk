// Package server_test contains unit/integration tests for product reviews service.
package server_test

import (
	"context"
	"os"
	"testing"

	apiv1 "github.com/eroshiva/cloudtalk/api/v1"
	"github.com/eroshiva/cloudtalk/internal/ent"
	"github.com/eroshiva/cloudtalk/internal/server"
	prs_testing "github.com/eroshiva/cloudtalk/pkg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	productName1        = "myAwesomeProduct#1"
	productName2        = "myAwesomeProduct#1"
	productDescription1 = "Product description #1"
	productDescription2 = "Product description #2"
	productPrice1       = "19.90"
	productPrice2       = "10.90"

	reviewer1Name     = "John"
	reviewer1LastName = "Doe"
	reviewer2Name     = "Theo"
	reviewer2LastName = "Walcott"
	reviewer3Name     = "OK"
	reviewer3LastName = "Guy"
	reviewer1Text     = "Product is good!"
	reviewer2Text     = "Product is OK."
	reviewer3Text     = "Product is bad!"
	reviewer1Rating   = 5
	reviewer2Rating   = 4
	reviewer3Rating   = 2
)

var (
	client     *ent.Client
	grpcClient apiv1.ProductReviewsServiceClient
)

func TestMain(m *testing.M) {
	var err error
	entClient, serverClient, wg, termChan, reverseProxyTermChan, err := prs_testing.SetupFull("", "")
	if err != nil {
		panic(err)
	}
	client = entClient
	grpcClient = serverClient

	// running tests
	code := m.Run()

	// all tests were run, stopping servers gracefully
	prs_testing.TeardownFull(client, wg, termChan, reverseProxyTermChan)
	os.Exit(code)
}

func TestCreateProduct(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), prs_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating request
	req := server.CreateProductRequest(productName1, productDescription1, productPrice1)

	// sending request to the server
	res, err := grpcClient.CreateProduct(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.GetProduct())
	t.Cleanup(func() {
		// cleaning up product resource at the end of the test
		_, err = grpcClient.DeleteProduct(ctx, server.DeleteProductRequest(res.GetProduct().GetId()))
		assert.NoError(t, err)
	})
	assert.Equal(t, productName1, res.GetProduct().GetName())
	assert.Equal(t, productDescription1, res.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, res.GetProduct().GetPrice())

	// retrieving product back by its ID
	product, err := grpcClient.GetProductByID(ctx, server.GetProductByIDRequest(res.GetProduct().GetId()))
	require.NoError(t, err)
	require.NotNil(t, product)
	require.NotNil(t, product.GetProduct())
	assert.Equal(t, productName1, product.GetProduct().GetName())
	assert.Equal(t, productDescription1, product.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, product.GetProduct().GetPrice())
}

func TestEditProduct(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), prs_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating request
	req := server.CreateProductRequest(productName1, productDescription1, productPrice1)

	// sending request to the server
	res, err := grpcClient.CreateProduct(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.GetProduct())
	t.Cleanup(func() {
		// cleaning up product resource at the end of the test
		_, err = grpcClient.DeleteProduct(ctx, server.DeleteProductRequest(res.GetProduct().GetId()))
		assert.NoError(t, err)
	})
	assert.Equal(t, productName1, res.GetProduct().GetName())
	assert.Equal(t, productDescription1, res.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, res.GetProduct().GetPrice())

	// retrieving product back by its ID
	product, err := grpcClient.GetProductByID(ctx, server.GetProductByIDRequest(res.GetProduct().GetId()))
	require.NoError(t, err)
	require.NotNil(t, product)
	require.NotNil(t, product.GetProduct())
	assert.Equal(t, productName1, product.GetProduct().GetName())
	assert.Equal(t, productDescription1, product.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, product.GetProduct().GetPrice())

	// changing product name
	updProduct, err := grpcClient.EditProduct(ctx, server.EditProductRequest(product.GetProduct().GetId(), productName2, "", ""))
	require.NoError(t, err)
	require.NotNil(t, updProduct)
	require.NotNil(t, updProduct.GetProduct())
	assert.Equal(t, productName2, updProduct.GetProduct().GetName())
	assert.Equal(t, productDescription1, updProduct.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, updProduct.GetProduct().GetPrice())
}

func TestGetProductByID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), prs_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating request
	req1 := server.CreateProductRequest(productName1, productDescription1, productPrice1)

	// sending request to the server
	res1, err := grpcClient.CreateProduct(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, res1)
	require.NotNil(t, res1.GetProduct())
	t.Cleanup(func() {
		// cleaning up product resource at the end of the test
		_, err = grpcClient.DeleteProduct(ctx, server.DeleteProductRequest(res1.GetProduct().GetId()))
		assert.NoError(t, err)
	})
	assert.Equal(t, productName1, res1.GetProduct().GetName())
	assert.Equal(t, productDescription1, res1.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, res1.GetProduct().GetPrice())

	// creating request
	req2 := server.CreateProductRequest(productName2, productDescription2, productPrice2)

	// sending request to the server
	res2, err := grpcClient.CreateProduct(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, res2)
	require.NotNil(t, res2.GetProduct())
	t.Cleanup(func() {
		// cleaning up product resource at the end of the test
		_, err = grpcClient.DeleteProduct(ctx, server.DeleteProductRequest(res2.GetProduct().GetId()))
		assert.NoError(t, err)
	})
	assert.Equal(t, productName2, res2.GetProduct().GetName())
	assert.Equal(t, productDescription2, res2.GetProduct().GetDescription())
	assert.Equal(t, productPrice2, res2.GetProduct().GetPrice())

	// retrieving product #1 back by its ID
	product1, err := grpcClient.GetProductByID(ctx, server.GetProductByIDRequest(res1.GetProduct().GetId()))
	require.NoError(t, err)
	require.NotNil(t, product1)
	require.NotNil(t, product1.GetProduct())
	assert.Equal(t, productName1, product1.GetProduct().GetName())
	assert.Equal(t, productDescription1, product1.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, product1.GetProduct().GetPrice())

	// retrieving product #2 back by its ID
	product2, err := grpcClient.GetProductByID(ctx, server.GetProductByIDRequest(res1.GetProduct().GetId()))
	require.NoError(t, err)
	require.NotNil(t, product2)
	require.NotNil(t, product2.GetProduct())
	assert.Equal(t, productName2, product2.GetProduct().GetName())
	assert.Equal(t, productDescription2, product2.GetProduct().GetDescription())
	assert.Equal(t, productPrice2, product2.GetProduct().GetPrice())
}

func TestListProducts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), prs_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating request
	req1 := server.CreateProductRequest(productName1, productDescription1, productPrice1)

	// sending request to the server
	res1, err := grpcClient.CreateProduct(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, res1)
	require.NotNil(t, res1.GetProduct())
	t.Cleanup(func() {
		// cleaning up product resource at the end of the test
		_, err = grpcClient.DeleteProduct(ctx, server.DeleteProductRequest(res1.GetProduct().GetId()))
		assert.NoError(t, err)
	})
	assert.Equal(t, productName1, res1.GetProduct().GetName())
	assert.Equal(t, productDescription1, res1.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, res1.GetProduct().GetPrice())

	// creating request
	req2 := server.CreateProductRequest(productName2, productDescription2, productPrice2)

	// sending request to the server
	res2, err := grpcClient.CreateProduct(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, res2)
	require.NotNil(t, res2.GetProduct())
	t.Cleanup(func() {
		// cleaning up product resource at the end of the test
		_, err = grpcClient.DeleteProduct(ctx, server.DeleteProductRequest(res2.GetProduct().GetId()))
		assert.NoError(t, err)
	})
	assert.Equal(t, productName2, res2.GetProduct().GetName())
	assert.Equal(t, productDescription2, res2.GetProduct().GetDescription())
	assert.Equal(t, productPrice2, res2.GetProduct().GetPrice())

	// listing all products in the system
	products, err := grpcClient.ListProducts(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.NotNil(t, products)
	// there is no reliable way to assert precise number of products in the system, because other tests concurrently add/remove other product resources.
	// at least two products added inside this test should be in the system in any case.
	assert.Greater(t, len(products.Products), 2)
}

func TestCreateReview(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), prs_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating request
	req1 := server.CreateProductRequest(productName1, productDescription1, productPrice1)

	// sending request to the server
	res1, err := grpcClient.CreateProduct(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, res1)
	require.NotNil(t, res1.GetProduct())
	t.Cleanup(func() {
		// cleaning up product resource at the end of the test
		_, err = grpcClient.DeleteProduct(ctx, server.DeleteProductRequest(res1.GetProduct().GetId()))
		assert.NoError(t, err)
	})
	assert.Equal(t, productName1, res1.GetProduct().GetName())
	assert.Equal(t, productDescription1, res1.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, res1.GetProduct().GetPrice())

	// creating review
	revReq1 := server.CreateReviewRequest(reviewer1Name, reviewer1LastName, reviewer1Text, reviewer1Rating)
	rev1, err := grpcClient.CreateReview(ctx, revReq1)
	require.NoError(t, err)
	require.NotNil(t, rev1)
	t.Cleanup(func() {
		// cleaning up review resource at the end of the test
		_, err = grpcClient.DeleteReview(ctx, server.DeleteReviewRequest(rev1.GetReview().GetId()))
		assert.NoError(t, err)
	})
	assert.Equal(t, reviewer1Name, rev1.GetReview().GetFirstName())
	assert.Equal(t, reviewer1LastName, rev1.GetReview().GetLastName())
	assert.Equal(t, reviewer1Text, rev1.GetReview().GetReviewText())
	assert.Equal(t, reviewer1Rating, rev1.GetReview().GetRating())

	// retrieving Product from the system - should contain one review
	product1, err := grpcClient.GetProductByID(ctx, server.GetProductByIDRequest(res1.GetProduct().GetId()))
	require.NoError(t, err)
	require.NotNil(t, product1)
	require.NotNil(t, product1.GetProduct())
	assert.Equal(t, productName1, product1.GetProduct().GetName())
	assert.Equal(t, productDescription1, product1.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, product1.GetProduct().GetPrice())
	assert.Equal(t, 1, len(product1.GetProduct().GetReviews()))

	// adding another review
	revReq2 := server.CreateReviewRequest(reviewer2Name, reviewer2LastName, reviewer2Text, reviewer2Rating)
	rev2, err := grpcClient.CreateReview(ctx, revReq2)
	require.NoError(t, err)
	require.NotNil(t, rev2)
	t.Cleanup(func() {
		// cleaning up review resource at the end of the test
		_, err = grpcClient.DeleteReview(ctx, server.DeleteReviewRequest(rev2.GetReview().GetId()))
		assert.NoError(t, err)
	})
	assert.Equal(t, reviewer2Name, rev2.GetReview().GetFirstName())
	assert.Equal(t, reviewer2LastName, rev2.GetReview().GetLastName())
	assert.Equal(t, reviewer2Text, rev2.GetReview().GetReviewText())
	assert.Equal(t, reviewer2Rating, rev2.GetReview().GetRating())

	// retrieving Product from the system - should contain two reviews
	product1, err = grpcClient.GetProductByID(ctx, server.GetProductByIDRequest(res1.GetProduct().GetId()))
	require.NoError(t, err)
	require.NotNil(t, product1)
	require.NotNil(t, product1.GetProduct())
	assert.Equal(t, productName1, product1.GetProduct().GetName())
	assert.Equal(t, productDescription1, product1.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, product1.GetProduct().GetPrice())
	assert.Equal(t, 2, len(product1.GetProduct().GetReviews()))
}

func TestGetReviewsByProductID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), prs_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating request
	req1 := server.CreateProductRequest(productName1, productDescription1, productPrice1)

	// sending request to the server
	res1, err := grpcClient.CreateProduct(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, res1)
	require.NotNil(t, res1.GetProduct())
	t.Cleanup(func() {
		// cleaning up product resource at the end of the test
		_, err = grpcClient.DeleteProduct(ctx, server.DeleteProductRequest(res1.GetProduct().GetId()))
		assert.NoError(t, err)
	})
	assert.Equal(t, productName1, res1.GetProduct().GetName())
	assert.Equal(t, productDescription1, res1.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, res1.GetProduct().GetPrice())

	// creating review
	revReq1 := server.CreateReviewRequest(reviewer1Name, reviewer1LastName, reviewer1Text, reviewer1Rating)
	rev1, err := grpcClient.CreateReview(ctx, revReq1)
	require.NoError(t, err)
	require.NotNil(t, rev1)
	t.Cleanup(func() {
		// cleaning up review resource at the end of the test
		_, err = grpcClient.DeleteReview(ctx, server.DeleteReviewRequest(rev1.GetReview().GetId()))
		assert.NoError(t, err)
	})

	// adding another review
	revReq2 := server.CreateReviewRequest(reviewer2Name, reviewer2LastName, reviewer2Text, reviewer2Rating)
	rev2, err := grpcClient.CreateReview(ctx, revReq2)
	require.NoError(t, err)
	require.NotNil(t, rev2)
	t.Cleanup(func() {
		// cleaning up review resource at the end of the test
		_, err = grpcClient.DeleteReview(ctx, server.DeleteReviewRequest(rev2.GetReview().GetId()))
		assert.NoError(t, err)
	})

	// adding one more review
	revReq3 := server.CreateReviewRequest(reviewer3Name, reviewer3LastName, reviewer3Text, reviewer3Rating)
	rev3, err := grpcClient.CreateReview(ctx, revReq3)
	require.NoError(t, err)
	require.NotNil(t, rev3)
	t.Cleanup(func() {
		// cleaning up review resource at the end of the test
		_, err = grpcClient.DeleteReview(ctx, server.DeleteReviewRequest(rev3.GetReview().GetId()))
		assert.NoError(t, err)
	})

	// retrieving all reviews by product ID
	reviews, err := grpcClient.GetReviewsByProductID(ctx, server.GetReviewsByProductIDRequest(res1.GetProduct().GetId()))
	require.NoError(t, err)
	require.NotNil(t, reviews)
	assert.Len(t, reviews, 3) // making sure there are precisely 3 reviews assigned to this product
}

func TestEditReview(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), prs_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating request
	req1 := server.CreateProductRequest(productName1, productDescription1, productPrice1)

	// sending request to the server
	res1, err := grpcClient.CreateProduct(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, res1)
	require.NotNil(t, res1.GetProduct())
	t.Cleanup(func() {
		// cleaning up product resource at the end of the test
		_, err = grpcClient.DeleteProduct(ctx, server.DeleteProductRequest(res1.GetProduct().GetId()))
		assert.NoError(t, err)
	})
	assert.Equal(t, productName1, res1.GetProduct().GetName())
	assert.Equal(t, productDescription1, res1.GetProduct().GetDescription())
	assert.Equal(t, productPrice1, res1.GetProduct().GetPrice())

	// creating review
	revReq1 := server.CreateReviewRequest(reviewer1Name, reviewer1LastName, reviewer1Text, reviewer1Rating)
	rev1, err := grpcClient.CreateReview(ctx, revReq1)
	require.NoError(t, err)
	require.NotNil(t, rev1)
	t.Cleanup(func() {
		// cleaning up review resource at the end of the test
		_, err = grpcClient.DeleteReview(ctx, server.DeleteReviewRequest(rev1.GetReview().GetId()))
		assert.NoError(t, err)
	})

	// editing the review's text
	editReq := server.EditReviewRequest(revReq1.GetReview().GetId(), "", "", reviewer2Text, reviewer2Rating)
	editResp, err := grpcClient.EditReview(ctx, editReq)
	require.NoError(t, err)
	require.NotNil(t, editResp)
	assert.Equal(t, reviewer1Name, editResp.GetReview().GetFirstName())
	assert.Equal(t, reviewer1LastName, editResp.GetReview().GetLastName())
	assert.Equal(t, reviewer2Text, editResp.GetReview().GetReviewText())
	assert.Equal(t, reviewer2Rating, editResp.GetReview().GetRating())
}
