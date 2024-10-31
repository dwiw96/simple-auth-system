BEGIN;

CREATE TABLE refresh_token_whitelist(
    id INT GENERATED ALWAYS AS IDENTITY
        CONSTRAINT pk_refresh_token_whitelist_id PRIMARY KEY,
    user_id INT NOT NULL
        CONSTRAINT uq_refresh_token_whitelist_user_id UNIQUE,
        CONSTRAINT fk_refresh_token_whitelist_user_id FOREIGN KEY (user_id)
            REFERENCES users(id),
    refresh_token UUID NOT NULL
        CONSTRAINT uq_refresh_token_whitelist_refresh_token UNIQUE,
    expires_at TIMESTAMP NOT NULL DEFAULT NOW() + INTERVAL '1 hour',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_refresh_token_whitelist_user_id ON refresh_token_whitelist(id);
CREATE INDEX ix_refresh_token_whitelist_refresh_token ON refresh_token_whitelist(refresh_token);
CREATE INDEX ix_refresh_token_whitelist_created_at ON refresh_token_whitelist(created_at);

COMMIT;
