#!/bin/sh

if [ -z "${GITHUB_USERNAME}" ]; then echo GITHUB_USERNAME must be set; exit 1; fi
if [ -z "${GITHUB_TOKEN}" ]; then echo GITHUB_TOKEN must be set; exit 1; fi
git config --global url."https://$GITHUB_USERNAME:$GITHUB_TOKEN@github.com/".insteadOf "https://github.com/"

if [ -f $PWD/bake-build ]; then
    echo "executing prebuilt bake file"
    exec $PWD/bake-build $@
else
    echo "executing mage"
    exec mage $@
fi

