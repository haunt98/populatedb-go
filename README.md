# populatedb

## Install

```sh
go install github.com/haunt98/populatedb-go/cmd/populatedb@latest
```

## Run

Example:

```sh
populatedb g --dialect "mysql" --url "root:@tcp(localhost:4000)/production" --table "production_2022" --number 10000000
```
