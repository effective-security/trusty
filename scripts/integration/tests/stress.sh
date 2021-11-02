#!/bin/bash
set -e

COUNT=$1
[ -z "$COUNT" ] && COUNT=100000

RANDOM=$$
CSR=/tmp/trusty_${RANDOM}

echo "CSR=$CSR"
echo "COUNT=$COUNT"

for ((n=0;n<COUNT;n++))
do
bin/trusty-tool --hsm-cfg=inmem csr create --plain-key --csr-profile=etc/dev/csr_profile/trusty_client.json  --out $CSR
bin/trustyctl -s https://localhost:7892 --timeout 3 --json \
    -c /tmp/trusty/certs/trusty_peer_ca.pem \
    -k /tmp/trusty/certs/trusty_peer_ca.key \
    -r /tmp/trusty/certs/trusty_root_ca.pem \
    ca sign --csr $CSR.csr --profile=server --label=$RANDOM 1> /dev/null

done

rm -rf $CSR*
