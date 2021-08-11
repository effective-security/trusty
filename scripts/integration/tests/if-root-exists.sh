#!/bin/bash
set -e

cmd="$*"

echo "*** trusty: waiting for certs..."

until [ -f /var/trusty/roots/trusty_root_ca.pem ];
do
  >&2 echo "trusty Root is unavailable - sleeping"
  sleep 1
done

>&2 echo "trusty is up - executing command:"
>&2 echo $cmd

exec $cmd