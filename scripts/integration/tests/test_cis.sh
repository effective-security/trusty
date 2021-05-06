#!/bin/bash
set -e

echo "*** trusty cis: checking against http endpoint: $TRUSTY_CIS_1"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_CIS_1 --timeout 3 cis roots

echo "*** trusty cis: checking against http endpoint: $TRUSTY_CIS_2"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_CIS_2 --timeout 3 --json cis roots
