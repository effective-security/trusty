#!/bin/bash

#
# gen_test_certs.sh
#   --out-dir {dir}         - specifies output folder
#   --out-prefix {prefix}   - specifies prefix for output files
#   --csr-dir {dir}         - specifies folder with CSR templates
#   --csr-prefix {prefix}   - specifies prefix for csr files
#   --hsm-confg             - specifies HSM provider file
#   --ca-config             - specifies CA configuration file
#   --root-ca {cert}        - specifies root CA certificate
#   --root-ca-key {key}     - specifies root CA key
#   --root                  - specifies if Root CA certificate and key should be generated
#   --ca                    - specifies if Level 1 CA certificate and key should be generated
#   --bundle                - specifies if Int CA Bundle should be created
#   --san                   - specifies SAN for server and peer certs
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
    --force)
    FORCE=YES
    shift # past argument
    ;;
    --bundle)
    BUNDLE=YES
    shift # past argument
    ;;
    --san)
    SAN="$2"
    shift # past argument
    shift # past value
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
[ -z "$ROOT_CA_KEY" ] && ROOT_CA_KEY=${OUT_DIR}/${OUT_PREFIX}root_ca-key.pem
[ -z "$SAN" ] && SAN=127.0.0.1

HOSTNAME=`hostname`

echo "OUT_DIR      = ${OUT_DIR}"
echo "OUT_PREFIX   = ${OUT_PREFIX}"
echo "CSR_DIR      = ${CSR_DIR}"
echo "CSR_PREFIX   = ${OUT_PREFIX}"
echo "CA_CONFIG    = ${CA_CONFIG}"
echo "HSM_CONFIG   = ${HSM_CONFIG}"
echo "BUNDLE       = ${BUNDLE}"
echo "FORCE        = ${FORCE}"
echo "SAN          = ${SAN}"
echo "ROOT_CA_CERT = $ROOT_CA_CERT"
echo "ROOT_CA_KEY  = $ROOT_CA_KEY"

if [[ "$ROOTCA" == "YES" && ("$FORCE" == "YES" || ! -f ${ROOT_CA_KEY}) ]]; then echo "*** generating ${ROOT_CA_CERT/.pem/''}"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} \
        csr gencert --self-sign \
        --ca-config=${CA_CONFIG} \
        --profile=SHAKEN_ROOT \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}root_ca.json \
        --key-label="${OUT_PREFIX}root_ca*" \
        --out ${ROOT_CA_CERT/.pem/''}
fi

if [[ "$CA" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}ca-key.pem) ]]; then
    echo "*** generating CA cert"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} \
        csr gencert \
        --ca-config=${CA_CONFIG} \
        --profile=SHAKEN_L1_CA \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}ca.json \
        --key-label="${OUT_PREFIX}ca*" \
        --ca-cert ${ROOT_CA_CERT} \
        --ca-key ${ROOT_CA_KEY} \
        --out ${OUT_DIR}/${OUT_PREFIX}ca
fi

if [[ "$BUNDLE" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}cabundle.pem) ]]; then
    echo "*** CA bundle"
    cat ${OUT_DIR}/${OUT_PREFIX}ca.pem >> ${OUT_DIR}/${OUT_PREFIX}cabundle.pem
fi
