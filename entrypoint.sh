#!/bin/sh

if [ -f $PWD/bake-build ]; then
    echo "executing prebuild bake file"
    exec $PWD/bake-build $@
else
    echo "executing mage"
    exec mage $@
fi

