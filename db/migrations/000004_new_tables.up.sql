-- Normally we would call this 'user', but then we would have to quote it everywhere
CREATE TABLE cms_user (
  id SERIAL PRIMARY KEY,
  body JSONB,
  insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- TODO: Is tit worth using a more specific index, e.g. on (Issuer,Subject)?
CREATE INDEX idx_user_body ON cms_user USING GIN (body);
GRANT SELECT, INSERT, UPDATE, DELETE ON cms_user TO ourroots;
GRANT USAGE, SELECT on SEQUENCE cms_user_id_seq to ourroots;

-- CREATE TABLE log (
--   id SERIAL PRIMARY KEY,
--   user_id INTEGER REFERENCES cms_user (id),
--   body JSONB,
--   insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--   last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );
-- CREATE INDEX idx_log_body ON log USING GIN (body);
-- GRANT SELECT, INSERT, UPDATE, DELETE ON log TO ourroots;
-- GRANT USAGE, SELECT on SEQUENCE log_id_seq to ourroots;

-- CREATE TABLE submission (
--   id SERIAL PRIMARY KEY,
--   user_id INTEGER REFERENCES cms_user (id),
--   collection_id INTEGER REFERENCES collection (id),
--   replaces_submission_id INTEGER REFERENCES submission (id),
--   body JSONB,
--   insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--   last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );
-- CREATE INDEX idx_submission_body ON submission USING GIN (body);
-- GRANT SELECT, INSERT, UPDATE, DELETE ON submission TO ourroots;
-- GRANT USAGE, SELECT on SEQUENCE submission_id_seq to ourroots;

-- CREATE TABLE record (
--   id SERIAL PRIMARY KEY,
--   submission_id INTEGER REFERENCES submission (id),
--   body JSONB,
--   insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--   last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );
-- CREATE INDEX idx_record_body ON record USING GIN (body);
-- GRANT SELECT, INSERT, UPDATE, DELETE ON record TO ourroots;
-- GRANT USAGE, SELECT on SEQUENCE record_id_seq to ourroots;

