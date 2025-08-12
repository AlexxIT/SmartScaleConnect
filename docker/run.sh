#!/bin/sh
set -e

if [ -f "/data/options.json" ]; then
  SLEEP=$(jq --raw-output ".sleep" /data/options.json)
fi

scaleconnect -i -r ${SLEEP:-24h}
