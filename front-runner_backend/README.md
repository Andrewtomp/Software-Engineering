# Documentation Notes

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