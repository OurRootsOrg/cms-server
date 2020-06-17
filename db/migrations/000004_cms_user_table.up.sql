-- Normally we would call this 'user', but then we would have to quote it everywhere
CREATE TABLE cms_user (
  id SERIAL PRIMARY KEY,
  body JSONB,
  insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- TODO: Is it worth using a more specific index, e.g. on (Issuer,Subject)?
CREATE INDEX idx_user_body ON cms_user USING GIN (body);
GRANT SELECT, INSERT, UPDATE, DELETE ON cms_user TO ourroots;
GRANT USAGE, SELECT on SEQUENCE cms_user_id_seq to ourroots;

