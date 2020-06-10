#!/bin/bash

set -e

echo
if [[ "$#" = 1 ]]; then
    case $1 in
        --gen-script) cat /home/beat/bake-default.sh; exit 0 ;;
        --gen-bin) mage -goos linux -compile bake-build; exit 0 ;;
    esac
fi


if [ -z "${GITHUB_TOKEN}" ]; then 
    echo "[WARN] GITHUB_TOKEN is not set, you won't be able to access private repos."
else
    git config --global url."https://$GITHUB_TOKEN@github.com/".insteadOf "https://github.com/"
fi

if [ -f $PWD/bake-build ]; then
    echo "executing prebuilt bake file"
    exec $PWD/bake-build $@
else
    echo "executing mage"
    exec mage $@
fi

