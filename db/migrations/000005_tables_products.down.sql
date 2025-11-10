    ALTER TABLE product_orders DROP CONSTRAINT "product_orders_id_order_fkey";
    ALTER TABLE product_orders DROP CONSTRAINT "product_orders_id_product_fkey";
    ALTER TABLE product DROP CONSTRAINT "product_id_product_images_fkey";
    ALTER TABLE product_categories DROP CONSTRAINT "product_categories_id_categories_fkey";
    ALTER TABLE product_categories DROP CONSTRAINT "product_categories_id_product_fkey";

    DROP TABLE IF EXISTS product_orders;
    DROP TABLE IF EXISTS product_categories;
    DROP TABLE IF EXISTS product;
    DROP TABLE IF EXISTS categories;
    DROP TABLE IF EXISTS product_images;