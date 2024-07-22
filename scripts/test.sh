#!/usr/bin/env bash

set -e

case "$1" in
  smoke)
    go test -run=Unit ./...
    ;;
  unit)
    go test -race -run=Unit ./...
    ;;
  integration)
    go test -race -run=Integration ./...
    ;;
  coverage)
    go test -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out
    ;;
  *)
    echo "Usage: ./test.sh [smoke|unit|integration|coverage]" && exit 1
    ;;
esac
