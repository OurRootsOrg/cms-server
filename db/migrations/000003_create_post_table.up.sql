CREATE TABLE IF NOT EXISTS post (
  id SERIAL PRIMARY KEY,
  collection_id INTEGER REFERENCES collection (id),
  body JSONB,
  insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
GRANT SELECT, INSERT, UPDATE, DELETE ON post TO ourroots;
GRANT USAGE, SELECT on SEQUENCE post_id_seq to ourroots;