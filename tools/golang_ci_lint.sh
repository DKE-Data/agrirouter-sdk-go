#!/bin/bash

GOLANGCI_LINT_VERSION=2.3.1

LINT_EXTRA_ARGS=""

if [[ "$1" == "all" ]] ; then 
  echo "Running golangci-lint on all files..."
else
  echo "Running golangci-lint on changed files only..."
  LINT_EXTRA_ARGS="--new-from-rev HEAD --whole-files"
fi

EXTRA_DOCKER_ARGS="-t -e GOLANGCI_LINT_CACHE=/app/.lint-cache/lint -e GOMODCACHE=/app/.lint-cache/gomod -e GOCACHE=/app/.lint-cache/gobuild" \
ENTRYPOINT=golangci-lint \
$(dirname -- "${BASH_SOURCE[0]}")/dockerized.sh golangci/golangci-lint:v${GOLANGCI_LINT_VERSION} run $LINT_EXTRA_ARGS

if [[ $? != 0 ]]; then
  echo "❌ Linting failed! Please fix errors before committing."
  exit 1
else
 echo "✅ ... lint done"
fi
