#!/bin/bash

set -e

BAKE_RUN_SCRIPT=run-bake.sh
BAKE_SCRIPT_LOCATIONS=( "." "./scripts" "./vendor/github.com/taxibeat/bake/scripts" )

for i in "${BAKE_SCRIPT_LOCATIONS[@]}"; do
  if [ ! -f "$i/$BAKE_RUN_SCRIPT" ]; then
    continue
  fi

  echo "Using bake run script at $i/$BAKE_RUN_SCRIPT"

  if [ ! -x "$i/$BAKE_RUN_SCRIPT" ]; then
      echo "Bake script is not executable, applying chmod..."
      chmod u+x "$i/$BAKE_RUN_SCRIPT"
  fi

  SCRIPT_FOUND=1
  "$i/$BAKE_RUN_SCRIPT" "$@"
  break
done

if [ -z $SCRIPT_FOUND ]; then
  echo Bake run script does not exist in any of these locations "${BAKE_SCRIPT_LOCATIONS[*]}"
  exit 1
fi
