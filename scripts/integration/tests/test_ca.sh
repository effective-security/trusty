#!/bin/bash
set -e

CA_FLAGS="-D --cfg /opt/trusty/etc/dev/trusty-config.yaml \
    -c /tmp/trusty/certs/trusty_peer_ca.pem \
    -k /tmp/trusty/certs/trusty_peer_ca.key \
    -r /tmp/trusty/certs/trusty_root_ca.pem"

echo "*** trusty ca: checking against http endpoint: $TRUSTY_CA_1"
/opt/trusty/bin/trustyctl $CA_FLAGS -s $TRUSTY_CA_1 --timeout 3 status
/opt/trusty/bin/trustyctl $CA_FLAGS -s $TRUSTY_CA_1 --timeout 3 caller
/opt/trusty/bin/trustyctl $CA_FLAGS -s $TRUSTY_CA_1 --timeout 3 ca issuers
#/opt/trusty/bin/trustyctl $CA_FLAGS -s $TRUSTY_CA_1 --timeout 3 ca issuer --label trusty.svc
#/opt/trusty/bin/trustyctl $CA_FLAGS -s $TRUSTY_CA_1 --timeout 3 ca issuer --label SHAKEN_CA
/opt/trusty/bin/trustyctl $CA_FLAGS -s $TRUSTY_CA_1 --timeout 3 ca profile server
/opt/trusty/bin/trustyctl $CA_FLAGS -s $TRUSTY_CA_1 --timeout 3 ca profile peer

echo "*** trusty ca: checking against http endpoint: $TRUSTY_CA_2"
/opt/trusty/bin/trustyctl $CA_FLAGS -s $TRUSTY_CA_2 --timeout 3 --o=json ca issuers
/opt/trusty/bin/trustyctl $CA_FLAGS -s $TRUSTY_CA_2 --timeout 3 ca profile DELEGATED_ICA

/opt/trusty/bin/hsm-tool --cfg=inmem csr create --plain-key --key-label=ci_test* --csr-profile=/opt/trusty/etc/dev/csr_profile/trusty_server.json --output /tmp/csr
/opt/trusty/bin/trustyctl -D --cfg /opt/trusty/etc/dev/trusty-config.yaml -s $TRUSTY_CA_2 --timeout 3 --o=json \
    -c /tmp/trusty/certs/trusty_peer_ca.pem \
    -k /tmp/trusty/certs/trusty_peer_ca.key \
    -r /tmp/trusty/certs/trusty_root_ca.pem \
    ca sign --csr /tmp/csr.csr --profile=server

#/opt/trusty/bin/trustyctl $CA_FLAGS -s $TRUSTY_CA_2 --timeout 3 --o=json ca certs --limit=1
