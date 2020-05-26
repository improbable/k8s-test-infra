#!/usr/bin/env bash
###
# Builds and pushes Prow containers to the Improbable GCR
###

set -e -o pipefail

# cd to the directory where this bash script is located at.
cd "$(dirname "$0")"
repo_root=$(dirname "$(pwd -P)")

export PROW_REPO_OVERRIDE="eu.gcr.io/windy-oxide-102215"
export DOCKER_REPO_OVERRIDE="${PROW_REPO_OVERRIDE}"
export EDGE_PROW_REPO_OVERRIDE="${PROW_REPO_OVERRIDE}"

docker build -t dockerized_tests dockerized_tests
docker run -e LOCAL_USER_ID="$(id -u)" \
  -v "${repo_root}":/repo:rw \
  --workdir=/repo \
  --entrypoint=/usr/local/bin/entrypoint.sh \
  -it \
  dockerized_tests \
  bash -c 'bazel \
    --bazelrc="/repo/improbable/bazelrc" \
    run \
    --config=imp-release \
    //improbable:improbable-push'
