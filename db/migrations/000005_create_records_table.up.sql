CREATE TABLE IF NOT EXISTS record (
  id SERIAL PRIMARY KEY,
  post_id INTEGER REFERENCES post (id),
  body JSONB,
  ix_hash TEXT DEFAULT '',
  insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_record_body ON record USING GIN (body);
CREATE INDEX idx_record_post ON record (post_id);
GRANT SELECT, INSERT, UPDATE, DELETE ON record TO ourroots;
GRANT USAGE, SELECT on SEQUENCE record_id_seq to ourroots;