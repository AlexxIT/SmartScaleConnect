#!/bin/sh
set -e

if [ -f "/data/options.json" ]; then
  SLEEP=$(jq --raw-output ".sleep" /data/options.json)
fi

while true; do
  scaleconnect
  sleep ${SLEEP:-24h}
done
