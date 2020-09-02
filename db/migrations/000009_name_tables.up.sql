CREATE TABLE IF NOT EXISTS givenname_variants (
    name TEXT PRIMARY KEY,
    variants JSONB,
    insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS surname_variants (
    name TEXT PRIMARY KEY,
    variants JSONB,
    insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
GRANT SELECT, INSERT, UPDATE, DELETE ON givenname_variants TO ourroots;
GRANT SELECT, INSERT, UPDATE, DELETE ON surname_variants TO ourroots;