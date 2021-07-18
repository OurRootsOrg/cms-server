-- database needs to be cleared out after this update
CREATE TABLE IF NOT EXISTS society (
    id  SERIAL PRIMARY KEY,
    body JSONB,
    insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
GRANT SELECT, INSERT, UPDATE, DELETE ON society TO ourroots;
GRANT USAGE, SELECT on SEQUENCE society_id_seq to ourroots;
INSERT INTO society (id) VALUES (0);

DROP TABLE IF EXISTS settings;

CREATE TABLE IF NOT EXISTS society_user (
    id  SERIAL PRIMARY KEY,
    body JSONB,
    user_id INTEGER REFERENCES cms_user (id) NOT NULL,
    society_id INTEGER REFERENCES society (id) NOT NULL,
    insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_society_user_society ON society_user (society_id, user_id);
CREATE INDEX idx_society_user_user ON society_user (user_id);
GRANT USAGE, SELECT on SEQUENCE society_user_id_seq to ourroots;
GRANT SELECT, INSERT, UPDATE, DELETE ON society_user TO ourroots;

CREATE TABLE IF NOT EXISTS invitation (
    id  SERIAL PRIMARY KEY,
    body JSONB,
    code TEXT NOT NULL,
    society_id INTEGER REFERENCES society (id) NOT NULL,
    insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_invitation_code ON invitation (code);
CREATE INDEX idx_invitation_society ON invitation (society_id);
GRANT USAGE, SELECT on SEQUENCE invitation_id_seq to ourroots;
GRANT SELECT, INSERT, UPDATE, DELETE ON invitation TO ourroots;

ALTER TABLE category ADD COLUMN IF NOT EXISTS society_id INTEGER REFERENCES society (id) NOT NULL DEFAULT 0;
CREATE INDEX idx_category_society ON category (society_id);

ALTER TABLE collection ADD COLUMN IF NOT EXISTS society_id INTEGER REFERENCES society (id) NOT NULL DEFAULT 0;
CREATE INDEX idx_collection_society ON collection (society_id);

ALTER TABLE post ADD COLUMN IF NOT EXISTS society_id INTEGER REFERENCES society (id) NOT NULL DEFAULT 0;
CREATE INDEX idx_post_society ON post (society_id);

ALTER TABLE record ADD COLUMN IF NOT EXISTS society_id INTEGER REFERENCES society (id) NOT NULL DEFAULT 0;
CREATE INDEX idx_record_society ON record (society_id);

ALTER TABLE record_household ADD COLUMN IF NOT EXISTS society_id INTEGER REFERENCES society (id) NOT NULL DEFAULT 0;
CREATE INDEX idx_record_household_society ON record_household (society_id);
