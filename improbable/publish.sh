#!/usr/bin/env bash
###
# Builds and pushes Prow containers to the Improbable GCR
###

set -e -o pipefail

# cd to the directory where this bash script is located at.
cd "$(dirname "$0")"
repo_root=$(dirname "$(pwd -P)")

GCLOUD_CONFIG_VOLUME_ARGS=()

if [[ -r "${HOME}/.config/gcloud/application_default_credentials.json" ]]; then
  GCLOUD_CONFIG_VOLUME_ARGS=(
    "--volume=${HOME}/.config/gcloud:/gcloud_config"
    "-e=CLOUDSDK_CONFIG=/gcloud_config"
    "-e=GOOGLE_APPLICATION_CREDENTIALS=/gcloud_config/application_default_credentials.json"
  )
fi

export PROW_REPO_OVERRIDE="eu.gcr.io/windy-oxide-102215"
export DOCKER_REPO_OVERRIDE="${PROW_REPO_OVERRIDE}"
export EDGE_PROW_REPO_OVERRIDE="${PROW_REPO_OVERRIDE}"

docker build -t dockerized_tests dockerized_tests
docker run -e LOCAL_USER_ID="$(id -u)" \
  -v "${repo_root}":/repo:rw \
  "${GCLOUD_CONFIG_VOLUME_ARGS[@]}" \
  --workdir=/repo \
  --entrypoint=/usr/local/bin/entrypoint.sh \
  -it \
  dockerized_tests \
  bash -c 'bazel \
    --bazelrc="improbable/bazelrc" \
    run \
    --config=imp-release \
    //improbable:improbable-push'
