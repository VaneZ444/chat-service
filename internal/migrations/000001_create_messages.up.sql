CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    author_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_messages_created_at ON messages(created_at);