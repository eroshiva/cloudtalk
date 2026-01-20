-- Modify "products" table
ALTER TABLE "products" ALTER COLUMN "average_rating" TYPE double precision USING (CASE WHEN average_rating ~ E'^\\d+(\\.\\d+)?$' THEN average_rating::double precision ELSE 0.0 END);
