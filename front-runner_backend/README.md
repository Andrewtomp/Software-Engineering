# Documentation Notes

## User Database

To create the database first you must get access to the postgresql shell:

```
psql -U johnny -d postgres
```

Replace `johnny` with the host system user (also need to change in the `.go` files).

Once in the shell run the following commands:

```
CREATE DATABASE users;

\c users

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(100) NOT NULL,
    business_name VARCHAR(100)
);

quit
```

## To Generate docs

```bash
swag init -g main.go -d cmd/main,internal/login --parseInternal
```

Additional `-d` directories will be added as we build more modules.

## To access the docs

Start the go server:

```golang
go run cmd/main/main.go
```

go to this webpage: https://localhost:8080/swagger/index.html