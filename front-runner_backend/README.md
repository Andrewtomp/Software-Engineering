# Documentation Notes

## User Database

To create the database first you must get access to the postgresql shell:

```
psql -U johnny -d postgres
```

Replace `johnny` with the host system user (make sure to update in the `.env` file).

Once in the shell run the following commands:

```
CREATE DATABASE frontrunner;

quit
```

gorm will create all the necessary tables for you.

## Build the Static Web Page

```bash
cd front-runner/

npm run build

cd ..
```

## Generate the Server Cert and the Storefront Encryption Key

```bash
cd front-runner_backend

bash ./generateCert.sh
```

## Start GO Server

Start the go server by running:

```golang
go run .
```

## To Access Page

go to this webpage: https://localhost:8080/

## To Generate docs

```bash
swag init --parseInternal
```

Additional `-d` directories will be added as we build more modules.

## To access the docs

go to this webpage: https://localhost:8080/swagger/index.html