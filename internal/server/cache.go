package server

import (
	"time"

	"github.com/eroshiva/cloudtalk/internal/ent"
	"github.com/maypok86/otter"
)

const defaultCachingTimeout = time.Minute

// Cache is a struct for caching operations.
type Cache struct {
	c otter.Cache[string, any]
}

// NewCache creates a new cache.
func NewCache() (*Cache, error) {
	c, err := otter.MustBuilder[string, any](1000).
		CollectStats().
		Cost(func(key string, value any) uint32 {
			return 1
		}).
		WithTTL(defaultCachingTimeout).
		Build()
	if err != nil {
		return nil, err
	}

	return &Cache{c: c}, nil
}

// GetProduct gets a product from the cache.
func (c *Cache) GetProduct(id string) (*ent.Product, bool) {
	p, ok := c.c.Get(id)
	if !ok {
		return nil, false
	}

	// casting back to original structure
	product, ok := p.(*ent.Product)
	if !ok {
		return nil, false
	}
	return product, ok
}

// SetProduct sets a product in the cache.
func (c *Cache) SetProduct(p *ent.Product) {
	c.c.Set(p.ID, p)
}

// DeleteProduct deletes a product from the cache.
func (c *Cache) DeleteProduct(id string) {
	c.c.Delete(id)
}

// GetProducts gets all products from the cache.
func (c *Cache) GetProducts() ([]*ent.Product, bool) {
	p, ok := c.c.Get("products")
	if !ok {
		return nil, false
	}

	// casting back to original structure
	products, ok := p.([]*ent.Product)
	if !ok {
		return nil, false
	}
	return products, true
}

// GetReviews gets reviews from the cache.
func (c *Cache) GetReviews(productID string) ([]*ent.Review, bool) {
	r, ok := c.c.Get("reviews:" + productID)
	if !ok {
		return nil, false
	}

	// casting back to original structure
	reviews, ok := r.([]*ent.Review)
	if !ok {
		return nil, false
	}
	return reviews, true
}

// SetReviews sets reviews in the cache.
func (c *Cache) SetReviews(productID string, reviews []*ent.Review) {
	c.c.Set("reviews:"+productID, reviews)
}

// DeleteReviews deletes reviews from the cache.
func (c *Cache) DeleteReviews(productID string) {
	c.c.Delete("reviews:" + productID)
}
