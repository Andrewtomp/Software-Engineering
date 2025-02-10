-- CREATE TYPE Login AS (
--     email VARCHAR(50),
--     password_hash VARCHAR(100)
-- );

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(100) NOT NULL,
    business_name VARCHAR(100)
);