#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
IMAGE=mongo:4.4

docker run \
  --rm \
  --network="crud" \
  -v ${DIR}/init.js:/app/init-rs.js \
  ${IMAGE} mongo --quiet "mongodb://mongo:mongo@mongo:27017" /app/init-rs.js
