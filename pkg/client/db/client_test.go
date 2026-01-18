// Package db_test contains complex unit tests for client.
package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/eroshiva/cloudtalk/internal/ent"
	"github.com/eroshiva/cloudtalk/pkg/client/db"
	prs_testing "github.com/eroshiva/cloudtalk/pkg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	productName1        = "myAwesomeProduct#1"
	productName2        = "myAwesomeProduct#1"
	productDescription1 = "Product description #1"
	productDescription2 = "Product description #2"
	productPrice1       = "19.90"
	productPrice2       = "10.90"
)

var (
	client *ent.Client
)

func TestMain(m *testing.M) {
	var err error
	client, err = prs_testing.Setup()
	if err != nil {
		panic(err)
	}
	code := m.Run()
	err = db.GracefullyCloseDBClient(client)
	if err != nil {
		panic(err)
	}
	os.Exit(code)
}

func TestProductCRUD(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), prs_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating product resource
	p, err := db.CreateProduct(ctx, client, productName1, productDescription1, productPrice1)
	require.NoError(t, err)
	require.NotNil(t, p)
	// cleaning up Product resource
	t.Cleanup(func() {
		err = db.DeleteProductByID(ctx, client, p.ID)
		assert.NoError(t, err)
	})

	// retrieving the same product and checking if all fields match
	retP, err := db.GetProductByID(ctx, client, p.ID)
	require.NoError(t, err)
	require.NotNil(t, retP)
	assert.Equal(t, productName1, retP.Name)
	assert.Equal(t, productDescription1, retP.Description)
	assert.Equal(t, productPrice1, retP.Price)

	// updating product description only
	updP, err := db.EditProduct(ctx, client, p.ID, "", productDescription2, "")
	require.NoError(t, err)
	require.NotNil(t, updP)
	assert.Equal(t, productName1, updP.Name)
	assert.Equal(t, productDescription2, updP.Description) // description is different one
	assert.Equal(t, productPrice1, updP.Price)

	// listing all products - there should be only one
	ps, err := db.ListProducts(ctx, client)
	require.NoError(t, err)
	require.NotNil(t, ps)
	assert.Len(t, ps, 1)
}
