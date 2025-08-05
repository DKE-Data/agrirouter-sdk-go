#!/bin/bash

set -e

EXTRA_DOCKER_ARGS=${EXTRA_DOCKER_ARGS:-""}

EXTRA_USER_ARGS=""
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    EXTRA_USER_ARGS="-u $(id -u ${USER}):$(id -g ${USER})"
fi

if [[ -n "$ENTRYPOINT" ]]; then
    ENTRYPOINT_ARG="--entrypoint ${ENTRYPOINT}"
fi

IMAGE=$1
shift 1

# needed for Git-Bash on Windows
export MSYS_NO_PATHCONV=1

docker run --rm -v $PWD:/app -w /app ${EXTRA_USER_ARGS} ${EXTRA_DOCKER_ARGS} ${ENTRYPOINT_ARG} "${IMAGE}" "$@"
