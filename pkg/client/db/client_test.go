// Package db_test contains complex unit tests for client.
package db_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/eroshiva/cloudtalk/internal/ent"
	"github.com/eroshiva/cloudtalk/internal/ent/product"
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

var client *ent.Client

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

func TestReviewCRUD(t *testing.T) {
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

	// creating first review (excellent)
	r1, err := db.CreateReview(ctx, client, reviewer1Name, reviewer1LastName, reviewer1Text, reviewer1Rating, p.ID)
	require.NoError(t, err)
	require.NotNil(t, r1)
	// cleaning up review
	t.Cleanup(func() {
		err = db.DeleteReviewByID(ctx, client, r1.ID, p.ID)
		assert.NoError(t, err)
	})
	assert.Equal(t, reviewer1Name, r1.FirstName)
	assert.Equal(t, reviewer1LastName, r1.LastName)
	assert.Equal(t, reviewer1Text, r1.ReviewText)
	assert.Equal(t, int32(reviewer1Rating), r1.Rating)

	// creating second review (OK)
	r2, err := db.CreateReview(ctx, client, reviewer2Name, reviewer2LastName, reviewer2Text, reviewer2Rating, p.ID)
	require.NoError(t, err)
	require.NotNil(t, r2)
	// cleaning up review
	t.Cleanup(func() {
		err = db.DeleteReviewByID(ctx, client, r2.ID, p.ID)
		assert.NoError(t, err)
	})
	assert.Equal(t, reviewer2Name, r2.FirstName)
	assert.Equal(t, reviewer2LastName, r2.LastName)
	assert.Equal(t, reviewer2Text, r2.ReviewText)
	assert.Equal(t, int32(reviewer2Rating), r2.Rating)

	// updating first review to be OK
	updR1, err := db.EditReview(ctx, client, r1.ID, "", "", reviewer2Text, reviewer2Rating)
	require.NoError(t, err)
	require.NotNil(t, updR1)
	assert.Equal(t, reviewer1Name, updR1.FirstName)
	assert.Equal(t, reviewer1LastName, updR1.LastName)
	assert.Equal(t, reviewer2Text, updR1.ReviewText)
	assert.Equal(t, int32(reviewer2Rating), updR1.Rating)

	// creating third review
	r3, err := db.CreateReview(ctx, client, reviewer3Name, reviewer3LastName, reviewer3Text, reviewer3Rating, p.ID)
	require.NoError(t, err)
	require.NotNil(t, r3)
	// cleaning up review
	t.Cleanup(func() {
		err = db.DeleteReviewByID(ctx, client, r3.ID, p.ID)
		assert.NoError(t, err)
	})
	assert.Equal(t, reviewer3Name, r3.FirstName)
	assert.Equal(t, reviewer3LastName, r3.LastName)
	assert.Equal(t, reviewer3Text, r3.ReviewText)
	assert.Equal(t, int32(reviewer3Rating), r3.Rating)

	// retrieve all reviews by Product ID
	rs, err := db.GetReviewsByProductID(ctx, client, p.ID)
	require.NoError(t, err)
	require.NotNil(t, rs)
	assert.Len(t, rs, 3)
}

// This new test verifies the database locking behavior. Two goroutines are trying to access the same resource.
// First one locks the resource in DB for time delta_t. Second one tries to access it.
// Second goroutine waits at least for delta_t time to get access to the resource.
// By measuring second goroutine execution time, we can assess whether locking has happened correctly.
// This test demonstrates the same mechanism as implemented in CRUD API for Review resource in this package.
func TestLockingOnConcurrentReviewCreation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), prs_testing.DefaultTestTimeout*2)
	t.Cleanup(cancel)

	// creating product
	p, err := db.CreateProduct(ctx, client, productName1, productDescription1, productPrice1)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = db.DeleteProductByID(ctx, client, p.ID)
		assert.NoError(t, err)
	})

	var wg sync.WaitGroup
	wg.Add(2)

	lockHeldChan := make(chan bool, 1)    // channel to signal that the lock is acquired
	releaseLockChan := make(chan bool, 1) // channel to signal the locker to commit

	lockingDuration := 200 * time.Millisecond // delta_t (see explanation above
	var waiterExecutionTime time.Duration     // variable to hold execution time of locked part in the goroutine

	// this goroutine starts a transaction, locks the Product resource, and waits for a signal to release it
	go func() {
		defer wg.Done()
		tx, err := client.Tx(ctx)
		require.NoError(t, err)
		// ensure rollback on error (if happens)
		defer tx.Rollback() //nolint:errcheck // this code doesn't go production

		// Acquire the lock
		_, err = tx.Product.Query().Where(product.ID(p.ID)).ForUpdate().Only(ctx)
		require.NoError(t, err)

		lockHeldChan <- true // signal that the lock is now held

		<-releaseLockChan // wait for the main test to tell us to release the lock

		// commit transaction
		require.NoError(t, tx.Commit())
	}()

	// this goroutine attempts to create a review, which should be blocked by other goroutine until it releases the lock
	go func() {
		defer wg.Done()
		<-lockHeldChan // wait until the lock is confirmed in the other goroutine

		startTime := time.Now()
		// this call will block until the transaction in other goroutine is committed!
		r, err := db.CreateReview(ctx, client, reviewer1Name, reviewer1LastName, reviewer1Text, reviewer1Rating, p.ID)
		waiterExecutionTime = time.Since(startTime)

		require.NoError(t, err)
		require.NotNil(t, r)
		t.Cleanup(func() {
			err = db.DeleteReviewByID(ctx, client, r.ID, p.ID)
			assert.NoError(t, err)
		})
	}()

	// wait a bit to ensure the locking goroutine has time to acquire the lock
	time.Sleep(lockingDuration)

	// signal to the locking goroutine to release the lock by committing its transaction
	releaseLockChan <- true

	// wait for both goroutines to complete
	wg.Wait()

	// asserting that waiting goroutine have taken at least as long as the lock was held.
	// waiting goroutine execution time should be >= locking duration to prove it was blocked
	assert.GreaterOrEqual(t, waiterExecutionTime, lockingDuration)
	t.Logf("Waiting goroutine was blocked for %v, proving lock contention.", waiterExecutionTime)

	// final state check to ensure data consistency after contention.
	retP, err := db.GetProductByID(ctx, client, p.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(retP.Edges.Reviews))                   // only 1 review was added
	assert.Equal(t, float64(reviewer1Rating), retP.AverageRating) // average rating must be equal to the only review's rating
}
