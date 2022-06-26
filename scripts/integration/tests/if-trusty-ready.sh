#!/bin/bash
set -e

if [[ -z "$TRUSTY_URL" ]]; then
    export TRUSTY_URL=http://168.138.72.101:7880
fi

echo "TRUSTY_URL: $TRUSTY_URL"

cmd="$*"

echo "*** trusty: waiting for server..."

until curl -f -v --fail -k $TRUSTY_URL/v1/status; do
  >&2 echo "trusty is unavailable $TRUSTY_URL - sleeping"
  sleep 6
done

>&2 echo "trusty is up - executing command:"
>&2 echo $cmd

exec $cmd