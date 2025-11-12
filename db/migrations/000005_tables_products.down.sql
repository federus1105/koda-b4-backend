ALTER TABLE product_orders DROP CONSTRAINT IF EXISTS "product_orders_id_order_fkey";
ALTER TABLE product_orders DROP CONSTRAINT IF EXISTS "product_orders_id_product_fkey";
ALTER TABLE product DROP CONSTRAINT IF EXISTS "product_id_product_images_fkey";
ALTER TABLE product_categories DROP CONSTRAINT IF EXISTS "product_categories_id_categories_fkey";
ALTER TABLE product_categories DROP CONSTRAINT IF EXISTS "product_categories_id_product_fkey";
ALTER TABLE product DROP CONSTRAINT IF EXISTS "product_id_size_product_fkey";
ALTER TABLE product DROP CONSTRAINT IF EXISTS "product_id_variant_product_fkey";

DROP TABLE IF EXISTS product_orders;
DROP TABLE IF EXISTS product_categories;
DROP TABLE IF EXISTS variant_product;
DROP TABLE IF EXISTS size_product;
DROP TABLE IF EXISTS product;
DROP TABLE IF EXISTS variants;
DROP TABLE IF EXISTS sizes;
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS categories;