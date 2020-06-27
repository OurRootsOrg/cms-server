ALTER TABLE collection ADD COLUMN IF NOT EXISTS category_id INTEGER REFERENCES category (id);
UPDATE collection SET category_id = (SELECT category_id FROM collection_category WHERE collection_id = id LIMIT 1);
DROP TABLE IF EXISTS collection_category;