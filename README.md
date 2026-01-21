Currently reviewer can add unlimited amounts of reviews to the product. It is not specified in the task if user can or cannot do this.

## Caching
Caching layer is implemented using `github.com/maypok86/otter`. 
It currently stores following data:
- product by ID.
- review list by product.

This is the simplest implementation of cache to satisfy the task.
Logic is following:
- whenever `Get` request happens, look into cache first.
  - if found in cache, return from cache.
  - if not found in cache, go to DB.
- whenever `Create` action happens:
  - in case of `Product`, simply add entry to cache.
  - in case of `Review`, invalidate whole review list cached for the specific product. It will be fetched again during `Get` operation.
- whenever `Edit` action happens, invalidate entry => it will be cached again on `Get` operation.
- whenever `Delete` action happens, simply remove the entry.

## Example commands

**CreateProduct**
```bash
curl -X POST "http://localhost:50052/v1/product/create" \
     -H "Content-Type: application/json" \
     -d '{ "product": { "name": "Example Product", "description": "This is an example product.", "price": "19.99" } }'
```

**GetProductByID**
```bash
curl -X GET "http://localhost:50052/v1/product/get/{product_id}"
```

**EditProduct**
```bash
curl -X PATCH "http://localhost:50052/v1/product/edit" \
     -H "Content-Type: application/json" \
     -d '{ "product": { "id": "{product_id}", "name": "Updated Product Name", "description": "Updated description.", "price": "29.99" } }'
```

**DeleteProduct**
```bash
curl -X DELETE "http://localhost:50052/v1/product/{product_id}"
```

**ListProducts**
```bash
curl -X GET "http://localhost:50052/v1/product/all"
```

**CreateReview**
```bash
curl -X POST "http://localhost:50052/v1/review/create" \
     -H "Content-Type: application/json" \
     -d '{ "review": { "product": { "id": "{product_id}" }, "first_name": "John", "last_name": "Doe", "review_text": "Great product!", "rating": 5 } }'
```

**GetReviewsByProductID**
```bash
curl -X GET "http://localhost:50052/v1/review/get/product/{product_id}"
```

**EditReview**
```bash
curl -X PATCH "http://localhost:50052/v1/review/edit" \
     -H "Content-Type: application/json" \
     -d '{ "review": { "id": "{review_id}", "first_name": "Jane", "last_name": "Doe", "review_text": "Good product, but a bit pricey.", "rating": 4 } }'
```

**DeleteReview**
```bash
curl -X DELETE "http://localhost:50052/v1/review/{review_id}"
```
