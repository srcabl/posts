CREATE TABLE IF NOT EXISTS links (
    uuid VARCHAR(36) NOT NULL UNIQUE,
    created_at INT(11) NOT NULL, -- UNIX time
    created_by_uuid VARCHAR(36) NOT NULL,
    updated_at INT(11), -- UNIX time
    updated_by_uuid VARCHAR(36),
    url VARCHAR(2048) NOT NULL,
    PRIMARY KEY(uuid)
);

CREATE TABLE IF NOT EXISTS posts (
    uuid VARCHAR(36) NOT NULL UNIQUE,
    created_at INT(11) NOT NULL,-- UNIX time
    created_by_uuid VARCHAR(36) NOT NULL,
    updated_at INT(11),-- UNIX time
    updated_by_uuid VARCHAR(36),
    user_uuid VARCHAR(36) NOT NULL,
    link_uuid VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    comment TEXT(100) NOT NULL,
    PRIMARY KEY(uuid),
    FOREIGN KEY(user_uuid) REFERENCES srcabl_users.users(uuid),
    FOREIGN KEY(link_uuid) REFERENCES srcabl_posts.links(uuid)
);

-- A link can have multiple primary souces (e.g. multiple authors)
-- This table stores the primary sources for each link
-- Each link has n entries for n primary sources
CREATE TABLE IF NOT EXISTS link_source_heads (
    link_uuid VARCHAR(36) NOT NULL,
    source_uuid VARCHAR(36) NOT NULL,
    PRIMARY KEY(link_uuid, source_uuid),
    FOREIGN KEY(link_uuid) REFERENCES srcabl_posts.links(uuid),
    FOREIGN KEY(source_uuid) REFERENCES srcabl_sources.sources(uuid)
);
