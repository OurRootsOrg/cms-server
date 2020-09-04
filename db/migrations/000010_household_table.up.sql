CREATE TABLE IF NOT EXISTS record_household (
    post_id INTEGER REFERENCES post (id),
    household_id TEXT NOT NULL,
    record_ids JSONB,
    insert_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, household_id)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON record_household TO ourroots;
