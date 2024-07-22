#!/usr/bin/env bash

set -e

case "${1}" in memcached | redis) ;; *) echo "Choose driver: [memcached, redis]" && exit 1 ;; esac

DRIVER="${1}"

memcached() {
  docker run -it -p 11211:11211 memcached:1.6.29
}

redis() {
  docker run -it -p 6379:6379 redis:7.2.5
}

echo "Starting driver: ${DRIVER}"

$"${DRIVER}"
