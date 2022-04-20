#!/bin/bash
set -e

echo "*** hsm-tool: softhsm"
export TOOL_FLAGS="-D --cfg /tmp/trusty/softhsm/unittest_hsm.json"

#echo "*** hsm-tool: hsm slots"
#/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm slots
echo "*** hsm-tool: hsm genkey RSA sign"
RSA_KEYID=$(/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm generate --purpose=sign --algo=RSA --size=2048 --label=ci_rsa_* | grep -o -E "id=(\w|[-_])+" | awk -F= '{print $2}')
echo "*** hsm-tool: hsm genkey RSA sign"
RSA_KEYID2=$(/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm generate --purpose=encrypt --algo=RSA --size=2048 --label=ci_rsa_* | grep -o -E "id=(\w|[-_])+" | awk -F= '{print $2}')
echo "*** hsm-tool: hsm genkey ECDSA"
ECC_KEYID=$(/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm generate --purpose=sign --algo=ECDSA --size=256 --label=ci_ecdsa_*  | grep -o -E "id=(\w|[-_])+" | awk -F= '{print $2}')
echo "*** hsm-tool: hsm list"
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm list
echo "*** hsm-tool: hsm info"
echo /opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm info --public $RSA_KEYID
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm info --public $RSA_KEYID
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm info --public $ECC_KEYID
echo "*** hsm-tool: hsm remove"
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm remove $RSA_KEYID
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm remove $RSA_KEYID2
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm remove $ECC_KEYID

echo "*** hsm-tool: csr create"
/opt/trusty/bin/hsm-tool $TOOL_FLAGS csr create \
        --plain-key \
        --csr-profile=/opt/trusty/etc/dev/csr_profile/trusty_l1_ca.json \
        --key-label="ci_csrtest_issuer1_ca*"
echo "*** hsm-tool: csr gencert"
/opt/trusty/bin/hsm-tool $TOOL_FLAGS csr gen-cert \
        --ca-config /opt/trusty/etc/dev/ca-config.bootstrap.yaml \
        --profile=L1_CA \
        --csr-profile=/opt/trusty/etc/dev/csr_profile/trusty_l1_ca.json \
        --key-label="ci_test_issuer1_ca*" \
        --ca-cert=/opt/trusty/etc/dev/roots/trusty_root_ca.pem \
        --ca-key=/opt/trusty/etc/dev/roots/trusty_root_ca.key \

echo "*** hsm-tool: local-kms"
export TOOL_FLAGS="-D --cfg /opt/trusty/tests/aws-test-kms.json"

#echo "*** hsm-tool: hsm slots"
#/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm slots
echo "*** hsm-tool: hsm genkey RSA"
RSA_KEYID=$(/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm generate --purpose=sign --algo=RSA --size=2048 --label=ci_sign_rsa_* | grep -o -E "id=(\w|[-_])+" | awk -F= '{print $2}')
RSA_KEYID2=$(/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm generate --purpose=encrypt --algo=RSA --size=2048 --label=ci_encr_rsa_* | grep -o -E "id=(\w|[-_])+" | awk -F= '{print $2}')
echo "*** hsm-tool: hsm genkey ECDSA"
ECC_KEYID=$(/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm generate --purpose=sign --algo=ECDSA --size=256 --label=ci_ecdsa_*  | grep -o -E "id=(\w|[-_])+" | awk -F= '{print $2}')
echo "*** hsm-tool: hsm list"
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm list
echo "*** hsm-tool: hsm info"
echo /opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm info --public $RSA_KEYID
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm info --public $RSA_KEYID
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm info --public $RSA_KEYID2
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm info --public $ECC_KEYID
echo "*** hsm-tool: hsm remove"
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm remove $RSA_KEYID
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm remove $RSA_KEYID2
/opt/trusty/bin/hsm-tool $TOOL_FLAGS hsm remove $ECC_KEYID
