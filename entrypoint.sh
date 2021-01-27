#!/bin/bash

set -e

if [ -z "${GITHUB_TOKEN}" ]; then echo GITHUB_TOKEN must be set; exit 1; fi
git config --global url."https://$GITHUB_TOKEN@github.com/".insteadOf "https://github.com/"

if [[ "$#" = 1 ]]; then
    case $1 in
        --gen-bin) mage -goos linux -compile magebin; exit 0 ;;
    esac
fi

if [ -f $PWD/magebin ]; then
    echo "Using prebuilt bake-build binary"
    exec $PWD/magebin $@
else
    exec mage $@
fi

