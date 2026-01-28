#!/bin/sh
set -e

# If arguments are provided, pass them directly
if [ $# -gt 0 ]; then
  exec scaleconnect "$@"
fi

# Otherwise, use default behavior
if [ -f "/data/options.json" ]; then
  SLEEP=$(jq --raw-output ".sleep" /data/options.json)
fi

exec scaleconnect -i -r ${SLEEP:-24h}
