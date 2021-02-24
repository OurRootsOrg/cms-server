ALTER TABLE category DROP COLUMN IF EXISTS society_id;
ALTER TABLE collection DROP COLUMN IF EXISTS society_id;
ALTER TABLE post DROP COLUMN IF EXISTS society_id;
ALTER TABLE record DROP COLUMN IF EXISTS society_id;
ALTER TABLE settings DROP COLUMN IF EXISTS society_id;
ALTER TABLE collection DROP COLUMN IF EXISTS society_id;
ALTER TABLE household DROP COLUMN IF EXISTS society_id;
DROP TABLE IF EXISTS invitation;
DROP TABLE IF EXISTS society_user;
DROP TABLE IF EXISTS society;
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY,
    body JSONB,
    insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
GRANT SELECT, INSERT, UPDATE, DELETE ON settings TO ourroots;
