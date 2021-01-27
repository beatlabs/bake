#!/bin/bash

set -e

image_name="taxibeat/bake"
image_tag="latest"

# GID to be added to user groups in the running container
# so that the user can interact with docker.
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
   docker_sock="/var/run/docker.sock"
   docker_gid=$(stat -c "%g" $docker_sock)
elif [[ "$OSTYPE" == "darwin"* ]]; then
   docker_sock="/var/run/docker.sock.raw"
   docker_gid=0
else
   echo "Unsupported OS"
   exit 1
fi

BAKE_SESSION_ID=$(uuidgen | cut -c 1,2,3 | tr "[:upper:]" "[:lower:]")
BAKE_NETWORK_ID=$(docker network create "${BAKE_SESSION_ID}")
printf "Bake Session ID: $BAKE_SESSION_ID\nBake Network ID: $BAKE_NETWORK_ID\n\n"

cleanup () {
    docker ps --format '{{.Names}}' | grep "^$BAKE_SESSION_ID-" | awk '{print $1}' | xargs -I {} docker rm -f {} > /dev/null
    # docker image list --format '{{.Repository}}' | grep "^$BAKE_SESSION_ID-" | awk '{print $1}' | xargs -I {} docker rmi -f {} > /dev/null
    docker image list --format '{{.Repository}}:{{.Tag}}' | grep ":$BAKE_SESSION_ID\$" | awk '{print $1}' | xargs -I {} docker rmi -f {} > /dev/null
    docker network rm "$BAKE_NETWORK_ID" > /dev/null
    echo "Bake cleanup complete"
}

if [[ "$SKIP_CLEANUP" != "1" ]]; then
trap cleanup EXIT
fi

# Detect TTY
[[ -t 1 ]] && tty='--tty'

docker run \
  --name "$BAKE_SESSION_ID-bake" \
  --network $BAKE_NETWORK_ID \
  $tty \
  --rm \
  --user $(id -u):$(id -g) \
  --group-add $docker_gid \
  --volume ${docker_sock}:/var/run/docker.sock \
  --volume "$PWD":/src \
  --workdir /src \
  --env BAKE_NETWORK_ID="$BAKE_NETWORK_ID" \
  --env BAKE_SESSION_ID="${BAKE_SESSION_ID}" \
  --env COVERALLS_TOKEN="$COVERALLS_TOKEN" \
  --env GITHUB_TOKEN="$GITHUB_TOKEN" \
  --env CONFLUENCE_USERNAME="$CONFLUENCE_USERNAME" \
  --env CONFLUENCE_PASSWORD="$CONFLUENCE_PASSWORD" \
  --env CONFLUENCE_BASEURL="$CONFLUENCE_BASEURL" \
  --env EXISTING_TESTSERVICE="$EXISTING_TESTSERVICE" \
  $image_name:$image_tag \
  $@
