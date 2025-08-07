#!/bin/sh
set -e

if [ -z "${SLEEP}" ]; then
    SLEEP=$(jq --raw-output '.sleep // "24h"' /data/options.json)
fi

while true; do
  scaleconnect
  sleep $SLEEP
done
