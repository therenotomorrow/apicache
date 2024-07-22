#!/usr/bin/env bash

set -e

gofmt -w -s .
golangci-lint run ./...
