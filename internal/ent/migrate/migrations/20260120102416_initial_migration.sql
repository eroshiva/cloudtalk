-- Create "products" table
CREATE TABLE "products" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NOT NULL,
  "price" character varying NOT NULL,
  "average_rating" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "reviews" table
CREATE TABLE "reviews" (
  "id" character varying NOT NULL,
  "first_name" character varying NOT NULL,
  "last_name" character varying NOT NULL,
  "review_text" character varying NOT NULL,
  "rating" integer NOT NULL,
  "product_reviews" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "reviews_products_reviews" FOREIGN KEY ("product_reviews") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
