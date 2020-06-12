#!/bin/bash

set -e

if [ -z "${GITHUB_TOKEN}" ]; then echo GITHUB_TOKEN must be set; exit 1; fi
git config --global url."https://$GITHUB_TOKEN@github.com/".insteadOf "https://github.com/"

if [[ "$#" = 1 ]]; then
    case $1 in
        --gen-script) cat /home/beat/bake-default.sh; exit 0 ;;
        --gen-bin) mage -goos linux -compile bake-build; exit 0 ;;
    esac
fi

if [ -f $PWD/bake-build ]; then
    echo "executing prebuilt bake file"
    exec $PWD/bake-build $@
else
    echo "executing mage"
    exec mage $@
fi

