#!/bin/bash

#
# gen_martini_certs.sh
#   --out-dir {dir}         - specifies output folder
#   --out-prefix {prefix}   - specifies prefix for output files
#   --csr-dir {dir}         - specifies folder with CSR templates
#   --csr-prefix {prefix}   - specifies prefix for csr files
#   --hsm-confg             - specifies HSM provider file
#   --crypto                - specifies additional HSM provider file
#   --ca-config             - specifies CA configuration file
#   --root-ca {cert}        - specifies root CA certificate
#   --root-ca-key {key}     - specifies root CA key
#   --root                  - specifies if Root CA certificate and key should be generated
#   --ca                    - specifies if CA certificate and key should be generated
#   --l1_ca                 - specifies if delegated Level 1 CA certificate and key should be generated
#   --bundle                - specifies if Int CA Bundle should be created
#   --force                 - specifies to force issuing the cert even if it exists
#

POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -o|--out-dir)
    OUT_DIR="$2"
    shift # past argument
    shift # past value
    ;;
    --out-prefix)
    OUT_PREFIX="$2"
    shift # past argument
    shift # past value
    ;;
    -o|--csr-dir)
    CSR_DIR="$2"
    shift # past argument
    shift # past value
    ;;
    --csr-prefix)
    CSR_PREFIX="$2"
    shift # past argument
    shift # past value
    ;;
    -c|--ca-config)
    CA_CONFIG="$2"
    shift # past argument
    shift # past value
    ;;
    --hsm-config)
    HSM_CONFIG="$2"
    shift # past argument
    shift # past value
    ;;
    --crypto)
    CRYPTO_PROV="--crypto-prov=$2"
    shift # past argument
    shift # past value
    ;;
    --root-ca)
    ROOT_CA_CERT="$2"
    shift # past argument
    shift # past value
    ;;
    --root-ca-key)
    ROOT_CA_KEY="$2"
    shift # past argument
    shift # past value
    ;;
    --root)
    ROOTCA=YES
    shift # past argument
    ;;
    --ca)
    CA=YES
    shift # past argument
    ;;
    --l1_ca)
    L1_CA=YES
    shift # past argument
    ;;
    --force)
    FORCE=YES
    shift # past argument
    ;;
    --bundle)
    BUNDLE=YES
    shift # past argument
    ;;
    *)
    echo "invalid flag $key: use --help to see the option"
    exit 1
esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

[ -z "$OUT_DIR" ] &&  echo "Specify --out-dir" && exit 1
[ -z "$CSR_DIR" ] &&  echo "Specify --csr-dir" && exit 1
[ -z "$CA_CONFIG" ] && echo "Specify --ca-config" && exit 1
[ -z "$HSM_CONFIG" ] && echo "Specify --hsm-config" && exit 1
[ -z "$ROOT_CA_CERT" ] && ROOT_CA_CERT=${OUT_DIR}/${OUT_PREFIX}root_ca.pem
[ -z "$ROOT_CA_KEY" ] && ROOT_CA_KEY=${OUT_DIR}/${OUT_PREFIX}root_ca.key

HOSTNAME=`hostname`

echo "OUT_DIR      = ${OUT_DIR}"
echo "OUT_PREFIX   = ${OUT_PREFIX}"
echo "CSR_DIR      = ${CSR_DIR}"
echo "CSR_PREFIX   = ${OUT_PREFIX}"
echo "CA_CONFIG    = ${CA_CONFIG}"
echo "HSM_CONFIG   = ${HSM_CONFIG}"
echo "CRYPTO_PROV  = ${CRYPTO_PROV}"
echo "BUNDLE       = ${BUNDLE}"
echo "FORCE        = ${FORCE}"
echo "ROOT_CA_CERT = $ROOT_CA_CERT"
echo "ROOT_CA_KEY  = $ROOT_CA_KEY"

if [[ "$ROOTCA" == "YES" && ("$FORCE" == "YES" || ! -f ${ROOT_CA_KEY}) ]]; then echo "*** generating ${ROOT_CA_CERT/.pem/''}"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} ${CRYPTO_PROV} \
        csr gencert --self-sign \
        --ca-config=${CA_CONFIG} \
        --profile=SHAKEN_ROOT \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}root_ca.json \
        --key-label="${OUT_PREFIX}root_ca*" \
        --out ${ROOT_CA_CERT/.pem/''}
fi

if [[ "$CA" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}ca.key) ]]; then
    echo "*** generating CA cert"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} ${CRYPTO_PROV} \
        csr gencert \
        --ca-config=${CA_CONFIG} \
        --profile=SHAKEN_CA \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}ca.json \
        --key-label="${OUT_PREFIX}ca*" \
        --ca-cert ${ROOT_CA_CERT} \
        --ca-key ${ROOT_CA_KEY} \
        --out ${OUT_DIR}/${OUT_PREFIX}ca

if [[ "$BUNDLE" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}cabundle.pem) ]]; then
    echo "*** CA bundle"
    cat ${OUT_DIR}/${OUT_PREFIX}ca.pem >> ${OUT_DIR}/${OUT_PREFIX}cabundle.pem
fi
fi

if [[ "$L1_CA" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}delegated_l1_ca.key) ]]; then
    echo "*** generating Delegated L1 CA cert"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} ${CRYPTO_PROV} \
        csr gencert \
        --ca-config=${CA_CONFIG} \
        --profile=DELEGATED_L1_CA \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}delegated_l1_ca.json \
        --key-label="${OUT_PREFIX}delegated_l1_ca*" \
        --ca-cert ${ROOT_CA_CERT} \
        --ca-key ${ROOT_CA_KEY} \
        --out ${OUT_DIR}/${OUT_PREFIX}delegated_l1_ca

if [[ "$BUNDLE" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}cabundle.pem) ]]; then
    echo "*** CA bundle"
    cat ${OUT_DIR}/${OUT_PREFIX}delegated_l1_ca.pem >> ${OUT_DIR}/${OUT_PREFIX}cabundle.pem
fi
fi
