#!/bin/bash -e

IMAGE=confa/migrate
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

docker build -t ${IMAGE} ${DIR}
docker run \
  --rm \
  -w /app \
  --network="crud" \
  -v ${DIR}/..:/app \
  ${IMAGE} migrate $@
