#!/bin/bash
set -e

echo "*** trusty ca: checking against http endpoint: $TRUSTY_SERVER_1"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_SERVER_1 --timeout 3 ca issuers

echo "*** trusty ca: checking against http endpoint: $TRUSTY_SERVER_2"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_SERVER_2 --timeout 3 --json ca issuers
