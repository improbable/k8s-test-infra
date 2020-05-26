#!/bin/bash
set -e -o pipefail

set -x

# cd to the directory where this bash script is located at.
cd "$(dirname "$0")"
repo_root=$(dirname "$(pwd -P)")

docker build -t dockerized_tests dockerized_tests
docker run -e LOCAL_USER_ID="$(id -u)" \
  -v "${repo_root}":/repo:rw \
  --workdir=/repo \
  --entrypoint=/usr/local/bin/entrypoint.sh \
  -it \
  dockerized_tests \
  bash -c 'bazel \
      test \
      --bazelrc="/repo/bazelrc" \
      --bazelrc="/repo/improbable/bazelrc" \
      --config=imp-ci \
      //prow/... \
      //ghproxy/...'
