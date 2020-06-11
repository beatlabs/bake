#!/bin/bash

set -e

image_name="taxibeat/bake"
image_tag="0.1.0"

DOCKER0_BRIDGE=172.17.0.1

# GID to be added to user groups in the running container
# so that the user can interact with docker.
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
   docker_gid=$(stat -c "%g" /var/run/docker.sock)
elif [[ "$OSTYPE" == "darwin"* ]]; then
   docker_gid=0
else
   echo "Unsupported OS"
   exit 1
fi

echo "Docker Group ID: $docker_gid"

RUN_ID=${RUN_ID:=$BUILD_NUMBER}
if [[ -z "$RUN_ID" ]]; then
    # Generate random 3 character alphanumeric string
    RUN_ID=$(uuidgen | cut -c 1,2,3)
fi

echo "Run ID: $RUN_ID"

# Force removal of containers and images.
cleanup () {
   docker ps --format '{{.Names}}' | grep "^$RUN_ID-" | awk '{print $1}' | xargs -I {} docker rm -f {}
   docker image list --format '{{.Repository}}' | grep "^$RUN_ID-" | awk '{print $1}' | xargs -I {} docker rmi -f {} > /dev/null
}
trap cleanup EXIT

echo "Starting run $RUN_ID"

# Detect TTY
[[ -t 1 ]] && t='-t'

docker run \
  --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $PWD:/src \
  -w /src \
  $t \
  --name "$RUN_ID-bake" \
  -e RUN_ID=$RUN_ID \
  -e CODECOV_TOKEN=$CODECOV_TOKEN \
  -e GITHUB_TOKEN=$GITHUB_TOKEN \
  -e CONFLUENCE_USERNAME=$CONFLUENCE_USERNAME \
  -e CONFLUENCE_PASSWORD=$CONFLUENCE_PASSWORD \
  -e CONFLUENCE_BASEURL=$CONFLUENCE_BASEURL \
  -e HOST_HOSTNAME=$DOCKER0_BRIDGE \
  -u $(id -u):$(id -g) \
  --group-add $docker_gid \
  $image_name:$image_tag \
  $@

