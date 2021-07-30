#!/bin/bash
set -e

echo "*** trusty: checking status with curl against CIS endpoint"
curl -f -v --fail $TRUSTY_CIS_1/v1/status

echo "*** trusty status: checking against http endpoint: $TRUSTY_CA_1"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_CA_1 --timeout 10 status
echo "*** trusty status: checking against http endpoint: $TRUSTY_WFE_1"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_WFE_1 --timeout 10 status
echo "*** trusty status: checking against http endpoint: $TRUSTY_WFE_2"
/opt/trusty/bin/trustyctl $TRUSTYCTL_FLAGS -s $TRUSTY_WFE_2 --timeout 10 status

echo "*** trusty: checking caller"

/opt/trusty/bin/trustyctl -V -D --cfg /opt/trusty/etc/dev/trusty-config.yaml -s $TRUSTY_WFE_2 \
    -c /tmp/trusty/certs/trusty_dev_peer_ca.pem \
    -k /tmp/trusty/certs/trusty_dev_peer_ca-key.pem \
    -r /tmp/trusty/certs/trusty_dev_root_ca.pem \
    caller | grep -c "Role | trusty-ca"

/opt/trusty/bin/trustyctl -V -D --cfg /opt/trusty/etc/dev/trusty-config.yaml -s $TRUSTY_WFE_2 \
    -c /tmp/trusty/certs/trusty_dev_peer_ra.pem \
    -k /tmp/trusty/certs/trusty_dev_peer_ra-key.pem \
    -r /tmp/trusty/certs/trusty_dev_root_ca.pem \
    caller | grep -c "Role | trusty-ra"

/opt/trusty/bin/trustyctl -V -D --cfg /opt/trusty/etc/dev/trusty-config.yaml -s $TRUSTY_WFE_2 \
    -c /tmp/trusty/certs/trusty_dev_peer_cis.pem \
    -k /tmp/trusty/certs/trusty_dev_peer_cis-key.pem \
    -r /tmp/trusty/certs/trusty_dev_root_ca.pem \
    caller | grep -c "Role | trusty-cis"

/opt/trusty/bin/trustyctl -V -D --cfg /opt/trusty/etc/dev/trusty-config.yaml -s $TRUSTY_WFE_2 \
    -c /tmp/trusty/certs/trusty_dev_peer_wfe.pem \
    -k /tmp/trusty/certs/trusty_dev_peer_wfe-key.pem \
    -r /tmp/trusty/certs/trusty_dev_root_ca.pem \
    caller | grep -c "Role | trusty-wfe"
