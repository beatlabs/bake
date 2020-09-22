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
#     export GITHUB_TOKEN=...
#     install.sh <VERSION> <FILE>
#
# for example to download the latest version:
#
#     install.sh latest bake_0.0.1-alpha.2_Linux_x86_64.tar.gz
#
# If your version/tag doesn't match, the script will exit with error.

set -x

REPO="taxibeat/bake"
GITHUB="https://api.github.com"

VERSION=$1
FILE=$2

if [ "$GITHUB_TOKEN" = "" ]; then
  echo "ERROR: missing GITHUB_TOKEN"
  exit 1
fi

if [ "$VERSION" = "" ]; then
  echo "ERROR: missing VERSION"
  exit 1
fi

if [ "$FILE" = "" ]; then
  echo "ERROR: missing FILE"
  exit 1
fi

if [ "$VERSION" = "latest" ]; then
  parser=".[0].assets | map(select(.name == \"$FILE\"))[0].id"
else
  parser=". | map(select(.tag_name == \"$VERSION\"))[0].assets | map(select(.name == \"$FILE\"))[0].id"
fi

asset_id=$(curl --header "Authorization: token $GITHUB_TOKEN" \
                --header "Accept: application/vnd.github.v3.raw" \
                --silent $GITHUB/repos/$REPO/releases | jq "$parser")
if [ "$asset_id" = "null" ]; then
  echo "ERROR: version not found $VERSION"
  exit 1
fi;

curl --header "Authorization: token $GITHUB_TOKEN" \
     --header 'Accept: application/octet-stream' \
     --silent --location --output "$FILE" \
     "$GITHUB/repos/$REPO/releases/assets/$asset_id"

tar --extract --file "$FILE" "${BAKE-bake}"
