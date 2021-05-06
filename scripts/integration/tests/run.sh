#!/bin/bash
#!/bin/sh
set -e

export TRUSTY_CIS_1=http://10.77.88.101:8080
export TRUSTY_CIS_2=http://10.77.88.102:8080
export TRUSTY_CIS_3=http://10.77.88.103:8080

export TRUSTY_SERVER_1=https://10.77.88.101:7891
export TRUSTY_SERVER_2=https://10.77.88.102:7891
export TRUSTY_SERVER_3=https://10.77.88.103:7891

export TRUSTY_CA_1=https://10.77.88.101:7892
export TRUSTY_CA_2=https://10.77.88.102:7892
export TRUSTY_CA_3=https://10.77.88.103:7892

export TRUSTY_ROOT=/var/trusty/roots/trusty_root_ca.pem

export TRUSTYCTL_FLAGS="-V -D --cfg /opt/trusty/etc/prod/trusty-config.yaml -r $TRUSTY_ROOT"

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

