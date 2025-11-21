#!/bin/sh
set -e
(
  cd "$(dirname "$0")"
  go build -o /tmp/build-http-server-go *.go
)
exec /tmp/build-http-server-go "$@"
