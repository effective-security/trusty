#!/bin/bash
#!/bin/sh
set -e

export TRUSTY_CIS_1=http://10.77.88.101:7880
export TRUSTY_CIS_2=http://10.77.88.102:7880

export TRUSTY_WFE_1=https://10.77.88.101:7891
export TRUSTY_WFE_2=https://10.77.88.102:7891

export TRUSTY_CA_1=https://10.77.88.101:7892
export TRUSTY_CA_2=https://10.77.88.102:7892

export TRUSTY_ROOT=/tmp/trusty/certs/trusty_dev_root_ca.pem

export TRUSTYCTL_FLAGS="-V -D --cfg /opt/trusty/etc/dev/trusty-config.yaml -r $TRUSTY_ROOT"

echo "Trusted anchors"
cp $TRUSTY_ROOT /etc/pki/ca-trust/source/anchors/
update-ca-trust extract
ls /etc/ssl/certs/
ls /etc/pki/ca-trust/source/anchors/

echo "*** Running trusty intergation tests using flags: $TRUSTYCTL_FLAGS"

echo "*** trusty: tools test"
source /opt/trusty/tests/test_tools.sh || (echo "tools test failed" && exit 1)

echo "*** trusty: status test"
source /opt/trusty/tests/test_status.sh || (echo "status test failed" && exit 1)

echo "*** trusty: CA test"
source /opt/trusty/tests/test_ca.sh || (echo "CA test failed" && exit 1)

echo "*** trusty: CIS test"
source /opt/trusty/tests/test_cis.sh || (echo "CIS test failed" && exit 1)

echo "*** trusty intergation tests succeeded"

