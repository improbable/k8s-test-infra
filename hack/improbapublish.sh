#!/usr/bin/env bash

###
# Builds Hook and pushes the container to the Improbable GCR
###

export PROW_REPO_OVERRIDE="eu.gcr.io/windy-oxide-102215"
export DOCKER_REPO_OVERRIDE="eu.gcr.io/windy-oxide-102215"

# bazel run //prow:release-push-hook --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64
bazel run \
  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
  //prow:improbable-push
