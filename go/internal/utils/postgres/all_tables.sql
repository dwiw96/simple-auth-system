BEGIN;
CREATE TABLE marital_status(
    id INT GENERATED ALWAYS AS IDENTITY
        CONSTRAINT pk_marital_status_id PRIMARY KEY,
    status VARCHAR(20) NOT NULL
        CONSTRAINT ck_marital_status_status CHECK (LENGTH(TRIM(status)) > 0),
        CONSTRAINT uq_marital_status_status UNIQUE(status)
);

CREATE TABLE users(
    id INT GENERATED ALWAYS AS IDENTITY
        CONSTRAINT pk_users_id PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL
        CONSTRAINT ck_users_first_name_length CHECK (LENGTH(TRIM(first_name)) > 0),
    middle_name VARCHAR(255) NULL,
    last_name VARCHAR(255) NOT NULL
        CONSTRAINT ck_users_last_name_length CHECK (LENGTH(TRIM(last_name)) > 0),
    email VARCHAR(50) NOT NULL
        CONSTRAINT uq_users_email UNIQUE,
        CONSTRAINT ck_users_email_length CHECK (LENGTH(TRIM(email)) >= 5),
    address VARCHAR(255) NOT NULL
        CONSTRAINT ck_users_address_length CHECK (LENGTH(TRIM(address)) > 3),
    gender VARCHAR (50) NOT NULL
        CONSTRAINT ck_users_gender_min CHECK (LENGTH(TRIM(gender)) > 0),
    marital_status_id INT NOT NULL,
        CONSTRAINT fk_users_marital_status_id FOREIGN KEY (marital_status_id)
            REFERENCES marital_status(id),
    hashed_password VARCHAR(255) NOT NULL
        CONSTRAINT ck_users_hashed_password_length CHECK (LENGTH(TRIM(hashed_password)) > 0),
        CONSTRAINT uq_users_hashed_password UNIQUE(hashed_password),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_users_first_name ON users(first_name);
CREATE INDEX ix_users_middle_name ON users(middle_name);
CREATE INDEX ix_users_last_name ON users(last_name);
CREATE INDEX ix_users_email ON users(email);
CREATE INDEX ix_users_address ON users(address);
CREATE INDEX ix_users_gender ON users(gender);
CREATE INDEX ix_users_created_at ON users(created_at);

CREATE TABLE emails_changes(
    id INT GENERATED ALWAYS AS IDENTITY
        CONSTRAINT pk_email_changes_id PRIMARY KEY,
    user_id INT NOT NULL,
        CONSTRAINT fk_email_change_user_id FOREIGN KEY (user_id)
            REFERENCES users(id),
    new_email VARCHAR(50) NULL,
    old_email VARCHAR(50) NOT NULL,
    email_change_at TIMESTAMP NOT NULL
);
COMMIT;