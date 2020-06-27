CREATE TABLE IF NOT EXISTS collection_category (
  collection_id INTEGER REFERENCES collection (id),
  category_id INTEGER REFERENCES category (id),
  primary key (collection_id, category_id)
);
CREATE INDEX idx_collection_category_collection ON collection_category (collection_id);
CREATE INDEX idx_collection_category_category ON collection_category (category_id);
GRANT SELECT, INSERT, UPDATE, DELETE ON collection_category TO ourroots;

INSERT INTO collection_category (collection_id, category_id) SELECT id, category_id from collection;
ALTER TABLE collection DROP COLUMN IF EXISTS category_id;