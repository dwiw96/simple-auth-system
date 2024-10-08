CREATE TABLE emails_changes(
    id INT GENERATED ALWAYS AS IDENTITY
        CONSTRAINT pk_emails_changes_id PRIMARY KEY,
    user_id INT NOT NULL
        CONSTRAINT fk_emails_changes_user_id FOREIGN KEY users(id),
    new_email VARCHAR(50) NULL,
    old_email VARCHAR(50) NOT NULL,
    emails_change_at TIMESTAMP NOT NULL,
);