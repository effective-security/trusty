#!/bin/bash
set -e

POSTGRES_HOST="$1"
POSTGRES_PORT="$2"
POSTGRES_USER="$3"
POSTGRES_PWD="$4"
shift # past host
shift # past port
shift # past user
shift # past pwd

cmd="$@"

echo "*** sql: waiting on cadb..."
sleep 3
until PGPASSWORD=$POSTGRES_PWD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -lqt | cut -d \| -f 1 | grep -qw cadb; do
  >&2 echo "cadb is unavailable $POSTGRES_HOST:$POSTGRES_PORT - sleeping"
  >&2 PGPASSWORD=$POSTGRES_PWD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -lqt
  sleep 3
done

>&2 echo "cadb is up - executing command:"
>&2 echo $cmd
exec $cmd