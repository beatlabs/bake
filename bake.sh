#!/bin/bash

set -e

image_name="taxibeat/bake"
image_tag="0.4.0"

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

NETWORK_ID=$(docker network create "${RUN_ID}"-bake)

echo "Network ID: $NETWORK_ID"

# Force removal of containers and images.
cleanup () {
   docker ps --format '{{.Names}}' | grep "^$RUN_ID-" | awk '{print $1}' | xargs -I {} docker rm -f {}
   docker image list --format '{{.Repository}}' | grep "^$RUN_ID-" | awk '{print $1}' | xargs -I {} docker rmi -f {} > /dev/null
   docker network rm "$NETWORK_ID" > /dev/null
}
trap cleanup EXIT

echo "Starting run $RUN_ID"

# Detect TTY
[[ -t 1 ]] && t='-t'

docker run \
  --network $NETWORK_ID \
  --rm \
  --volume /var/run/docker.sock:/var/run/docker.sock \
  --volume "$PWD":/src \
  --workdir /src \
  $t \
  --name "$RUN_ID-bake" \
  --env RUN_ID="$RUN_ID" \
  --env CODECOV_TOKEN="$CODECOV_TOKEN" \
  --env GITHUB_TOKEN="$GITHUB_TOKEN" \
  --env CONFLUENCE_USERNAME="$CONFLUENCE_USERNAME" \
  --env CONFLUENCE_PASSWORD="$CONFLUENCE_PASSWORD" \
  --env CONFLUENCE_BASEURL="$CONFLUENCE_BASEURL" \
  --env NETWORK_ID="$NETWORK_ID" \
  -u $(id -u):$(id -g) \
  --group-add $docker_gid \
  $image_name:$image_tag \
  $@

