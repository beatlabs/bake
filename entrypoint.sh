#!/bin/sh

if [ -f $PWD/mage-build ]; then
    exec $PWD/mage-build $@
else
    exec mage $@
fi

