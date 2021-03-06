CREATE TABLE IF NOT EXISTS collection (
  id SERIAL PRIMARY KEY,
  category_id INTEGER REFERENCES category (id),
  body JSONB,
  insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
GRANT SELECT, INSERT, UPDATE, DELETE ON collection TO ourroots;
GRANT USAGE, SELECT on SEQUENCE collection_id_seq to ourroots;