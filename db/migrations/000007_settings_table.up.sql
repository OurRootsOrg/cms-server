CREATE TABLE IF NOT EXISTS settings (
  id INTEGER PRIMARY KEY,
  body JSONB,
  insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
GRANT SELECT, INSERT, UPDATE, DELETE ON settings TO ourroots;