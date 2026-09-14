CREATE TABLE comments (
    id            BIGSERIAL PRIMARY KEY,
    blog_id       BIGINT NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    parent_id     BIGINT REFERENCES comments(id) ON DELETE CASCADE,
    is_anonymous  BOOLEAN NOT NULL DEFAULT FALSE,
    author_name   TEXT,
    author_email  TEXT,
    author_ip     TEXT,
    content       TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_blog_id ON comments(blog_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_id);

-- Reuses update_updated_at(), already defined by migration 001.
CREATE TRIGGER comments_updated_at
BEFORE UPDATE ON comments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
