To create the database first you must get access to the postgresql shell:

```
psql -U johnny -d postgres
```

Once in the shell run the following commands:

```
CREATE DATABASE users;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(100) NOT NULL
);
```