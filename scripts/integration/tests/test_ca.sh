#!/bin/bash
set -e

echo "*** trusty ca: checking against http endpoint: $TRUSTY_CA_1"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_CA_1 --timeout 3 status
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_CA_1 --timeout 3 ca issuers
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_CA_1 --timeout 3 ca profile --name server
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_CA_1 --timeout 3 ca profile --name peer --issuer TrustyCA

echo "*** trusty ca: checking against http endpoint: $TRUSTY_CA_2"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_CA_2 --timeout 3 --json ca issuers

/opt/trusty/bin/trusty-tool --hsm-cfg=inmem csr create --plain-key --csr-profile=/opt/trusty/etc/prod/csr/trusty_server.json --out /tmp/csr
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_CA_2 --timeout 3 --json ca sign --csr /tmp/csr.csr --profile=server
