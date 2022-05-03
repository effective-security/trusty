#!/bin/bash
set -e

echo "*** trusty: checking status with curl against CIS endpoint"
curl -f -v --fail $TRUSTY_CIS_1/v1/status

echo "*** trusty status: checking against http endpoint: $TRUSTY_CA_1"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_CA_1 --timeout 10 status
echo "*** trusty status: checking against http endpoint: $TRUSTY_CA_2"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_CA_2 --timeout 10 status

echo "*** trusty: checking caller"

/opt/trusty/bin/trustyctl -V -D --cfg /opt/trusty/etc/dev/trusty-config.yaml -s $TRUSTY_CA_2 \
    -c /tmp/trusty/certs/trusty_peer_ca.pem \
    -k /tmp/trusty/certs/trusty_peer_ca.key \
    -r /tmp/trusty/certs/trusty_root_ca.pem \
    caller | grep -c "| trusty-ca"

/opt/trusty/bin/trustyctl -V -D --cfg /opt/trusty/etc/dev/trusty-config.yaml -s $TRUSTY_CA_2 \
    -c /tmp/trusty/certs/trusty_peer_ra.pem \
    -k /tmp/trusty/certs/trusty_peer_ra.key \
    -r /tmp/trusty/certs/trusty_root_ca.pem \
    caller | grep -c "| trusty-ra"

/opt/trusty/bin/trustyctl -V -D --cfg /opt/trusty/etc/dev/trusty-config.yaml -s $TRUSTY_CA_1 \
    -c /tmp/trusty/certs/trusty_peer_cis.pem \
    -k /tmp/trusty/certs/trusty_peer_cis.key \
    -r /tmp/trusty/certs/trusty_root_ca.pem \
    caller | grep -c "| trusty-cis"

/opt/trusty/bin/trustyctl -V -D --cfg /opt/trusty/etc/dev/trusty-config.yaml -s $TRUSTY_CA_2 \
    -c /tmp/trusty/certs/trusty_peer_wfe.pem \
    -k /tmp/trusty/certs/trusty_peer_wfe.key \
    -r /tmp/trusty/certs/trusty_root_ca.pem \
    caller | grep -c "| trusty-wfe"

echo "*** trusty: checking caller done"