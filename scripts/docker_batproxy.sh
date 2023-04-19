#!/usr/bin/env bash
set -eo pipefail

# variable set for image
label=v0.0.8
service=docker.io/ashertz/batproxy
container_name="${service}:${label}"
sqlite_db=$(pwd)/batproxy.db
unix_sock=/var/run/batproxy
listen="unix:///var/run/batproxy/batproxy.sock"

# variable set for docker
network=eci

function batproxy_run() {
    if [ ! -f "${sqlite_db}" ]; then
        >&2 echo "sqlite db doesn't existed, try to run 'scripts/init_db.sh' first"
        exit 1
    fi

    docker run -d \
        --network ${network} \
        --name batproxy \
        -v "${sqlite_db}":/batproxy.db \
        -v ${unix_sock}:/var/run/batproxy \
        ${container_name} \
        --listen ${listen} \
        "${@}"

}

function batproxy_proxy() {
    docker exec -it -e BATPROXY_BASE_URL=${listen} batproxy /batproxy "${@}"
}

function batproxy_stop() {
    docker rm -f batproxy

    unix_sock_file=${unix_sock}/batproxy.sock

    rm -f ${unix_sock_file}

    echo "unix file cleaned"
}

case $1 in
  "run")
  batproxy_run "${@}"
  ;;
  "proxy")
  batproxy_proxy "${@}"
  ;;
  "stop")
  batproxy_stop "${@}"
  ;;
  *)
  >&2 echo "select one of subcommand [run, proxy, stop]"
  exit 1
esac