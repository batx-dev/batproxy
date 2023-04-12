#!/usr/bin/env bash
set -x 
set -eo pipefail

function init_mysql() {
  MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD:=password}
  MYSQL_DATABASE=${MYSQL_DATABASE:=batproxy}
  MYSQL_HOST=${MYSQL_HOST:=127.0.0.1}
  MYSQL_PORT=${MYSQL_PORT:=3306}

  # Allow to skip Docker if a dockerized MySQL database is already running
  if [[ -z "${SKIP_DOCKER}" ]]
  then
      # Launch mysql using Docker
      docker run \
          --name mysql \
          -e MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD} \
          -p "${MYSQL_PORT}":3306 \
          -d mysql:5.7
  fi

  # Keep pinging MySQL until it's ready to accept commands
  until MYSQL_PWD="${MYSQL_ROOT_PASSWORD}" mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "root" -e "CREATE DATABASE IF NOT EXISTS ${MYSQL_DATABASE}"
  do
      >&2 echo "MySQL is still unavailable - sleeping"
      sleep 1
  done

  >&2 echo "MySQL is up and running on port ${DB_PORT}!"

  MYSQL_PWD="${MYSQL_ROOT_PASSWORD}" mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "root" -D "${MYSQL_DATABASE}" -e "source migrations/t_bat_proxy_mysql.sql"
  >&2 echo "MySQL has been migrated, ready to go!"
}

function init_sqlite3() {
  SQLITE_DATABASE=${SQLITE_DATABASE:=batproxy.db}

  sqlite3 "${SQLITE_DATABASE}" < migrations/t_bat_proxy_sqlite3.sql
}

case $1 in
  "mysql")
  init_mysql
  ;;
  "sqlite")
  init_sqlite3
  ;;
  *)
  >&2 echo "select one of database drive [mysql, sqlite]"
  exit 1
esac




