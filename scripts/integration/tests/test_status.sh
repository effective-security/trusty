#!/bin/bash
set -e

echo "*** trusty status: checking against http endpoint: $TRUSTY_SERVER_1"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_SERVER_1 --timeout 10 status
echo "*** trusty status: checking against http endpoint: $TRUSTY_SERVER_2"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_SERVER_2 --timeout 10 status
#/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_SERVER_3_HTTP --timeout 10 status

echo "*** trusty: checking status with curl against CIS endpoint"
curl -f -v --fail $TRUSTY_CIS_1/v1/status
curl -f -v --fail $TRUSTY_CA_1/v1/status
