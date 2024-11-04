CREATE TABLE
    sessions (
        id SERIAL PRIMARY KEY,
        user_id INT UNIQUE,
        token_hash TEXT UNIQUE NOT NULL
    );

INSERT INTO
    sessions (user_id, token_hash)
VALUES
    (1, "hashed_token")
SELECT