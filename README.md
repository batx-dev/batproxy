# batproxy

A HTTP reverse proxy, but connecting to the backend through SSH.

## Develop Guidance

1. Init database by [init_db](scripts/init_db.sh)
    ```shell
    # this command will generate sqlite3 db file `batproxy.db`
    $./scripts/init_db.sh sqlite
    ```
   
2. Move `batproxy.db` file to [.batproxy](./.batproxy) directory

3. Run program by use `make run`

## Deploy Guidance

1. Prepare `batproxy.db` for store reverse proxy rule

2. Run `batproxy` by [docker_batproxy.sh](scripts/docker_batproxy.sh)

## Use Guidance

* As deployer, want to know more deploy options, use `batproxy run -h`

* As api user, know more by [api.md](docs/api.md)

## Checklist
- [X] Sqlite
- [ ] Mysql

