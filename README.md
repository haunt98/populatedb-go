# populatedb

Quickly populate database with fake data for testing only.

Currently support:

- mysql:
  - varchar
  - bigint
  - int
  - timestamp
  - json

## Install

```sh
go install github.com/haunt98/populatedb-go/cmd/populatedb@latest
```

## Run

Example:

```sh
populatedb g --dialect "mysql" --url "root:@tcp(localhost:4000)/production" --table "production_2022" --number 10000000
```

## Contribute

Feel free to ask or implement feature yourself :)

## Roadmap

- [ ] Support more database (postgres, sqlite, ...)
- [ ] Support more database type

## Thanks

- [k1LoW/tbls](https://github.com/k1LoW/tbls)
- [urfave/cli](https://github.com/urfave/cli)
