#!/usr/bin/env bash
#
# This script downloads an asset from latest or specific Github release of a
# private repo.
#
# PREREQUISITES
#
# curl, jq
#
# USAGE
#
#     GITHUB_TOKEN=... $BAKE_VERSION=... install.sh
#
# OR
#
#     GITHUB_TOKEN=... install.sh <$BAKE_VERSION>
#

set -x

REPO="taxibeat/bake"
GITHUB="https://api.github.com"

if [ "$BAKE_VERSION" = "" ]; then
  BAKE_VERSION=$1
fi

if [ "$GITHUB_TOKEN" = "" ]; then
  echo "ERROR: missing GITHUB_TOKEN"
  exit 1
fi

if [ "$BAKE_VERSION" = "" ]; then
  echo "ERROR: missing VERSION"
  exit 1
fi

FILE="bake-$BAKE_VERSION-Linux-x86_64.tar.gz"

parser=". | map(select(.tag_name == \"$BAKE_VERSION\"))[0].assets | map(select(.name == \"$FILE\"))[0].id"
asset_id=$(curl --header "Authorization: token $GITHUB_TOKEN" \
                --header "Accept: application/vnd.github.v3.raw" \
                --silent $GITHUB/repos/$REPO/releases | jq "$parser")
if [ "$asset_id" = "null" ]; then
  echo "ERROR: bake version not found $BAKE_VERSION"
  exit 1
fi;

curl --header "Authorization: token $GITHUB_TOKEN" \
     --header 'Accept: application/octet-stream' \
     --silent --location --output "$FILE" \
     "$GITHUB/repos/$REPO/releases/assets/$asset_id"

tar --extract --file "$FILE" "${BAKE-bake}"
