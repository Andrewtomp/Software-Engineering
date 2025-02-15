# Documentation Notes

## User Database

To create the database first you must get access to the postgresql shell:

```
psql -U johnny -d postgres
```

Replace `johnny` with the host system user (also need to change in the `.go` files).

Once in the shell run the following commands:

```
CREATE DATABASE frontrunner;

\c frontrunner

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(100) NOT NULL,
    business_name VARCHAR(100)
);

quit
```

## Start GO Server

Start the go server by running:

```golang
go run front-runner_backend/main.go
```

## To Generate docs

```bash
swag init --parseInternal
```

Additional `-d` directories will be added as we build more modules.

## To access the docs

go to this webpage: https://localhost:8080/swagger/index.html