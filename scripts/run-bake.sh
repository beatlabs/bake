#!/bin/bash

set -e

# parse parameters in order to pipe specific flags directly to docker
DOCKER_ENV=""
while test $# -gt 1; do
  if [ "$1" == "--env" ]; then
    DOCKER_ENV="${DOCKER_ENV} $1 $2"
    shift
    shift
  else
    break
  fi
done

image_name="ghcr.io/beatlabs/bake"
image_tag=`cat go.mod | grep github.com/beatlabs/bake | cut -f2 -d"v"`

# "module" value means that this script is used in the Bake repository itself
if [ "$image_tag" == "module github.com/beatlabs/bake" ]; then
  echo "Setting bake image tag to latest"
  image_tag="latest"
fi

if ! [[ `cat ${HOME}/.docker/config.json | grep ghcr.io` || `docker-credential-desktop list | grep ghcr.io` ]]; then
  echo "docker config not found for ghcr.io, please log in"
fi

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
  logsdir='.bake-container-logs'
  mkdir -p $logsdir
  docker ps --format '{{.Names}}' | grep "^$BAKE_SESSION_ID-" | awk '{print $1}' | xargs -I {} sh -c "docker logs {} > $logsdir/{}.log 2>&1"
  docker ps --format '{{.Names}}' | grep "^$BAKE_SESSION_ID-" | awk '{print $1}' | xargs -I {} docker rm -f {} > /dev/null
  docker image list --format '{{.Repository}}:{{.Tag}}' | grep ":$BAKE_SESSION_ID\$" | awk '{print $1}' | xargs -I {} docker rmi -f {} > /dev/null
  docker network rm "$BAKE_NETWORK_ID" > /dev/null
  rm -f docker/component/.bakesession > /dev/null
  rm -f test/.bakesession > /dev/null
  echo "Bake cleanup complete"
}

if [[ "$SKIP_CLEANUP" != "1" ]]; then
trap cleanup EXIT
fi

# Detect TTY
[[ -t 1 ]] && tty='--tty'

echo "GITHUB_TOKEN=${GITHUB_TOKEN}"

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
  --env GITHUB_TOKEN="${GITHUB_TOKEN}" \
  --env CONFLUENCE_USERNAME="$CONFLUENCE_USERNAME" \
  --env CONFLUENCE_PASSWORD="$CONFLUENCE_PASSWORD" \
  --env CONFLUENCE_BASEURL="$CONFLUENCE_BASEURL" \
  --env GITHUB_ACTIONS="$GITHUB_ACTIONS" \
  --env BAKE_HOST_PATH="${PWD}" \
  --env BAKE_PUBLISH_PORTS="true" \
  $DOCKER_ENV \
  $image_name:$image_tag \
  $@
