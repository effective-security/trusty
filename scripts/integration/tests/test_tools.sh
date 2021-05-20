#!/bin/bash
set -e

echo "*** trusty tool: softhsm"
export TOOL_FLAGS="-V -D --hsm-cfg /tmp/trusty/softhsm/unittest_hsm.json"

echo "*** trusty tool: hsm slots"
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm slots
echo "*** trusty tool: hsm genkey RSA sign"
RSA_KEYID=$(/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm genkey --purpose=sign --alg=RSA --size=2048 --label=ci_rsa_* | grep -o -E "id=(\w|[-_])+" | awk -F= '{print $2}')
echo "*** trusty tool: hsm genkey RSA sign"
RSA_KEYID2=$(/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm genkey --purpose=encrypt --alg=RSA --size=2048 --label=ci_rsa_* | grep -o -E "id=(\w|[-_])+" | awk -F= '{print $2}')
echo "*** trusty tool: hsm genkey ECDSA"
ECC_KEYID=$(/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm genkey --purpose=sign --alg=ECDSA --size=256 --label=ci_ecdsa_*  | grep -o -E "id=(\w|[-_])+" | awk -F= '{print $2}')
echo "*** trusty tool: hsm lskey"
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm lskey
echo "*** trusty tool: hsm keyinfo"
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm keyinfo --public --id=$RSA_KEYID
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm keyinfo --public --id=$ECC_KEYID
echo "*** trusty tool: hsm rmkey"
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm rmkey --id=$RSA_KEYID
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm rmkey --id=$RSA_KEYID2
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm rmkey --id=$ECC_KEYID

echo "*** trusty tool: csr create"
/opt/trusty/bin/trusty-tool $TOOL_FLAGS csr create \
        --plain-key \
        --csr-profile=/opt/trusty/etc/prod/csr/trusty_issuer1_ca.json \
        --key-label="ci_csrtest_issuer1_ca*"
echo "*** trusty tool: csr gencert"
/opt/trusty/bin/trusty-tool $TOOL_FLAGS csr gencert \
        --ca-config /opt/trusty/etc/prod/ca-config.bootstrap.yaml \
        --profile=L1_CA \
        --csr-profile=/opt/trusty/etc/prod/csr/trusty_issuer1_ca.json \
        --key-label="ci_test_issuer1_ca*" \
        --ca-cert=/var/trusty/roots/trusty_root_ca.pem \
        --ca-key=/var/trusty/roots/trusty_root_ca-key.pem \


echo "*** trusty tool: local-kms"
export TOOL_FLAGS="-V -D --hsm-cfg /opt/trusty/tests/aws-test-kms.json"

echo "*** trusty tool: hsm slots"
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm slots
echo "*** trusty tool: hsm genkey RSA"
RSA_KEYID=$(/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm genkey --purpose=sign --alg=RSA --size=2048 --label=ci_sign_rsa_* | grep -o -E "id=(\w|[-_])+" | awk -F= '{print $2}')
RSA_KEYID2=$(/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm genkey --purpose=encrypt --alg=RSA --size=2048 --label=ci_encr_rsa_* | grep -o -E "id=(\w|[-_])+" | awk -F= '{print $2}')
echo "*** trusty tool: hsm genkey ECDSA"
ECC_KEYID=$(/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm genkey --purpose=sign --alg=ECDSA --size=256 --label=ci_ecdsa_*  | grep -o -E "id=(\w|[-_])+" | awk -F= '{print $2}')
echo "*** trusty tool: hsm lskey"
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm lskey
echo "*** trusty tool: hsm keyinfo"
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm keyinfo --public --id=$RSA_KEYID
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm keyinfo --public --id=$RSA_KEYID2
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm keyinfo --public --id=$ECC_KEYID
echo "*** trusty tool: hsm rmkey"
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm rmkey --id=$RSA_KEYID
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm rmkey --id=$RSA_KEYID2
/opt/trusty/bin/trusty-tool $TOOL_FLAGS hsm rmkey --id=$ECC_KEYID
