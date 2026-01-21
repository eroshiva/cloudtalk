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
