# batproxy

A HTTP reverse proxy, but connecting to the backend through SSH.

## Develop Guidance

1. init database by [init_db](scripts/init_db.sh)
    ```shell
    # this command will generate sqlite3 db file `batproxy.db`
    $./scripts/init_db.sh sqlite
    ```
   
2. move `batproxy.db` file to [.batproxy](./.batproxy) directory

3. Run program by use `make run`

## Checklist
- [X] Sqlite
- [ ] Mysql

