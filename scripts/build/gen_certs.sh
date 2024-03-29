#!/bin/bash

#
# gen_certs.sh
#   --output-dir {dir}         - specifies output folder
#   --output-prefix {prefix}   - specifies prefix for output files
#   --csr-dir {dir}         - specifies folder with CSR templates
#   --csr-prefix {prefix}   - specifies prefix for csr files
#   --key-label {prefix}    - specifies prefix for key label, by default empty
#   --hsm-confg             - specifies HSM provider file
#   --crypto                - specifies additional HSM provider file
#   --ca-config             - specifies CA configuration file
#   --root-ca {cert}        - specifies root CA certificate
#   --root-ca-key {key}     - specifies root CA key
#   --root                  - specifies if Root CA certificate and key should be generated
#   --ca1                   - specifies if Level 1 CA certificate and key should be generated
#   --ca2                   - specifies if Level 2 CA certificate and key should be generated
#   --server                - specifies if server TLS certificate and key should be generated
#   --client                - specifies if client certificate and key should be generated
#   --peer                  - specifies if peer certificate and key should be generated
#   --bundle                - specifies if Int CA Bundle should be created
#   --san                   - specifies SAN for server and peer certs
#   --force                 - specifies to force issuing the cert even if it exists
#   --verbose               - specifies to enable verbose logs
#

POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -o|--output-dir|--out-dir)
    OUT_DIR="$2"
    shift # past argument
    shift # past value
    ;;
    --output-prefix|--out-prefix)
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
    -l|--key-label)
    KEY_LABEL="$2"
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
    CRYPTO_PROV="--crypto=$2"
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
    --ca1)
    CA1=YES
    shift # past argument
    ;;
    --ca2)
    CA2=YES
    shift # past argument
    ;;
    --server)
    SERVER=YES
    shift # past argument
    ;;
    --admin)
    ADMIN=YES
    shift # past argument
    ;;
    --client)
    CLIENT=YES
    shift # past argument
    ;;
    --peers|--peer)
    PEERS=YES
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
    --san|--SAN)
    SAN="$2"
    shift # past argument
    shift # past value
    ;;
    --verbose)
    FLAGS="-D"
    shift # past argument
    ;;
    *)
    echo "invalid flag $key: use --help to see the option"
    exit 1
esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

[ -z "$OUT_DIR" ] &&  echo "Specify --output-dir" && exit 1
[ -z "$CSR_DIR" ] &&  echo "Specify --csr-dir" && exit 1
[ -z "$CA_CONFIG" ] && echo "Specify --ca-config" && exit 1
[ -z "$HSM_CONFIG" ] && echo "Specify --hsm-config" && exit 1
[ -z "$ROOT_CA_CERT" ] && ROOT_CA_CERT=${OUT_DIR}/${OUT_PREFIX}root_ca.pem
[ -z "$ROOT_CA_KEY" ] && ROOT_CA_KEY=${OUT_DIR}/${OUT_PREFIX}root_ca.key
[ -z "$SAN" ] && SAN=127.0.0.1

HOSTNAME=`hostname`

echo "FLAGS        = ${FLAGS}"
echo "OUT_DIR      = ${OUT_DIR}"
echo "OUT_PREFIX   = ${OUT_PREFIX}"
echo "CSR_DIR      = ${CSR_DIR}"
echo "CSR_PREFIX   = ${OUT_PREFIX}"
echo "CA_CONFIG    = ${CA_CONFIG}"
echo "HSM_CONFIG   = ${HSM_CONFIG}"
echo "CRYPTO_PROV  = ${CRYPTO_PROV}"
echo "KEY_LABEL    = ${KEY_LABEL}"
echo "BUNDLE       = ${BUNDLE}"
echo "FORCE        = ${FORCE}"
echo "SAN          = ${SAN}"
echo "ROOT_CA_CERT = ${ROOT_CA_CERT}"
echo "ROOT_CA_KEY  = ${ROOT_CA_KEY}"

if [[ "$FLAGS" == "-D" ]]; then echo "*** hsm-tool "
    hsm-tool --version; 
fi

if [[ "$ROOTCA" == "YES" && ("$FORCE" == "YES" || ! -f ${ROOT_CA_KEY}) ]]; then echo "*** generating ${ROOT_CA_CERT/.pem/''}"
    hsm-tool ${FLAGS} \
        --cfg ${HSM_CONFIG} ${CRYPTO_PROV} \
        csr gen-cert --self-sign \
        --ca-config ${CA_CONFIG} \
        --profile ROOT \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}root_ca.yaml \
        --key-label "${KEY_LABEL}${OUT_PREFIX}root_ca*" \
        --output ${ROOT_CA_CERT/.pem/''}
fi

if [[ "$CA1" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}l1_ca.key) ]]; then
    echo "*** generating L1 CA cert"
    hsm-tool ${FLAGS} \
        --cfg ${HSM_CONFIG} ${CRYPTO_PROV} \
        csr gen-cert \
        --ca-config ${CA_CONFIG} \
        --profile L1_CA \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}l1_ca.yaml \
        --key-label "${KEY_LABEL}${OUT_PREFIX}l1_ca*" \
        --ca-cert ${ROOT_CA_CERT} \
        --ca-key ${ROOT_CA_KEY} \
        --output ${OUT_DIR}/${OUT_PREFIX}l1_ca
fi

if [[ "$CA2" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}l2_ca.key) ]]; then
    echo "*** generating L2 CA cert"
    hsm-tool ${FLAGS} \
        --cfg=${HSM_CONFIG}  ${CRYPTO_PROV} \
        csr gen-cert \
        --ca-config ${CA_CONFIG} \
        --profile L2_CA \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}l2_ca.yaml \
        --key-label "${KEY_LABEL}${OUT_PREFIX}l2_ca*" \
        --ca-cert ${OUT_DIR}/${OUT_PREFIX}l1_ca.pem \
        --ca-key ${OUT_DIR}/${OUT_PREFIX}l1_ca.key \
        --output ${OUT_DIR}/${OUT_PREFIX}l2_ca
fi

if [[ "$BUNDLE" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}cabundle.pem) ]]; then
    echo "*** CA bundle"
    if [[ -f ${OUT_DIR}/${OUT_PREFIX}l2_ca.pem ]]; then
        cat ${OUT_DIR}/${OUT_PREFIX}l2_ca.pem > ${OUT_DIR}/${OUT_PREFIX}cabundle.pem
    fi
    if [[ -f ${OUT_DIR}/${OUT_PREFIX}l1_ca.pem ]]; then
        cat ${OUT_DIR}/${OUT_PREFIX}l1_ca.pem >> ${OUT_DIR}/${OUT_PREFIX}cabundle.pem
    fi
fi

if [[ "$ADMIN" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}admin.key) ]]; then
    echo "*** generating admin cert"
    hsm-tool ${FLAGS} \
        --cfg ${HSM_CONFIG}  ${CRYPTO_PROV} \
        csr gen-cert --plain-key \
        --ca-config=${CA_CONFIG} \
        --profile client \
        --ca-cert ${OUT_DIR}/${OUT_PREFIX}l2_ca.pem \
        --ca-key ${OUT_DIR}/${OUT_PREFIX}l2_ca.key \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}admin.yaml \
        --key-label "${KEY_LABEL}${OUT_PREFIX}admin*" \
        --output ${OUT_DIR}/${OUT_PREFIX}admin

        cat ${OUT_DIR}/${OUT_PREFIX}cabundle.pem >> ${OUT_DIR}/${OUT_PREFIX}admin.pem
fi

if [[ "$SERVER" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}server.key) ]]; then
    echo "*** generating server cert"
    hsm-tool ${FLAGS} \
        --cfg ${HSM_CONFIG}  ${CRYPTO_PROV} \
        csr gen-cert --plain-key \
        --ca-config ${CA_CONFIG} \
        --profile server \
        --ca-cert ${OUT_DIR}/${OUT_PREFIX}l2_ca.pem \
        --ca-key ${OUT_DIR}/${OUT_PREFIX}l2_ca.key \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}server.yaml \
        --san localhost,${SAN},${HOSTNAME} \
        --key-label="${KEY_LABEL}${OUT_PREFIX}server*" \
        --output ${OUT_DIR}/${OUT_PREFIX}server

        cat ${OUT_DIR}/${OUT_PREFIX}cabundle.pem >> ${OUT_DIR}/${OUT_PREFIX}server.pem
fi

if [[ "$CLIENT" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}client.key) ]]; then
    echo "*** generating client cert"
    hsm-tool ${FLAGS} \
        --cfg ${HSM_CONFIG}  ${CRYPTO_PROV} \
        csr gen-cert --plain-key \
        --ca-config ${CA_CONFIG} \
        --profile client \
        --ca-cert ${OUT_DIR}/${OUT_PREFIX}l2_ca.pem \
        --ca-key ${OUT_DIR}/${OUT_PREFIX}l2_ca.key \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}client.yaml \
        --san spiffe://trusty/client \
        --key-label "${KEY_LABEL}${OUT_PREFIX}client*" \
        --output ${OUT_DIR}/${OUT_PREFIX}client

    cat ${OUT_DIR}/${OUT_PREFIX}cabundle.pem >> ${OUT_DIR}/${OUT_PREFIX}client.pem
fi

if [[ "$PEERS" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}peer_ca.key) ]]; then
    echo "*** generating peer_ca cert"
    hsm-tool ${FLAGS} \
        --cfg ${HSM_CONFIG}  ${CRYPTO_PROV} \
        csr gen-cert --plain-key \
        --ca-config ${CA_CONFIG} \
        --profile peer \
        --ca-cert ${OUT_DIR}/${OUT_PREFIX}l2_ca.pem \
        --ca-key ${OUT_DIR}/${OUT_PREFIX}l2_ca.key \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}peer_ca.yaml \
        --san localhost,${SAN},${HOSTNAME},spiffe://trusty/ca \
        --key-label "${KEY_LABEL}${OUT_PREFIX}peer_ca*" \
        --output ${OUT_DIR}/${OUT_PREFIX}peer_ca

    cat ${OUT_DIR}/${OUT_PREFIX}cabundle.pem >> ${OUT_DIR}/${OUT_PREFIX}peer_ca.pem
fi

if [[ "$PEERS" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}peer_ra.key) ]]; then
    echo "*** generating peer_ra cert"
    hsm-tool ${FLAGS} \
        --cfg ${HSM_CONFIG}  ${CRYPTO_PROV} \
        csr gen-cert --plain-key \
        --ca-config ${CA_CONFIG} \
        --profile peer \
        --ca-cert ${OUT_DIR}/${OUT_PREFIX}l2_ca.pem \
        --ca-key ${OUT_DIR}/${OUT_PREFIX}l2_ca.key \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}peer_ra.yaml \
        --san localhost,${SAN},${HOSTNAME},spiffe://trusty/ra \
        --key-label "${KEY_LABEL}${OUT_PREFIX}peer_ra*" \
        --output ${OUT_DIR}/${OUT_PREFIX}peer_ra

    cat ${OUT_DIR}/${OUT_PREFIX}cabundle.pem >> ${OUT_DIR}/${OUT_PREFIX}peer_ra.pem
fi

if [[ "$PEERS" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}peer_cis.key) ]]; then
    echo "*** generating peer_cis cert"
    hsm-tool ${FLAGS} \
        --cfg ${HSM_CONFIG}  ${CRYPTO_PROV} \
        csr gen-cert --plain-key \
        --ca-config ${CA_CONFIG} \
        --profile peer \
        --ca-cert ${OUT_DIR}/${OUT_PREFIX}l2_ca.pem \
        --ca-key ${OUT_DIR}/${OUT_PREFIX}l2_ca.key \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}peer_cis.yaml \
        --san localhost,${SAN},${HOSTNAME},spiffe://trusty/cis \
        --key-label "${KEY_LABEL}${OUT_PREFIX}peer_cis*" \
        --output ${OUT_DIR}/${OUT_PREFIX}peer_cis

    cat ${OUT_DIR}/${OUT_PREFIX}cabundle.pem >> ${OUT_DIR}/${OUT_PREFIX}peer_cis.pem
fi

if [[ "$PEERS" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${OUT_PREFIX}peer_wfe.key) ]]; then
    echo "*** generating peer_wfe cert"
    hsm-tool ${FLAGS} \
        --cfg ${HSM_CONFIG}  ${CRYPTO_PROV} \
        csr gen-cert --plain-key \
        --ca-config=${CA_CONFIG} \
        --profile peer \
        --ca-cert ${OUT_DIR}/${OUT_PREFIX}l2_ca.pem \
        --ca-key ${OUT_DIR}/${OUT_PREFIX}l2_ca.key \
        --csr-profile ${CSR_DIR}/${CSR_PREFIX}peer_wfe.yaml \
        --san localhost,${SAN},${HOSTNAME},spiffe://trusty/wfe \
        --key-label "${KEY_LABEL}${OUT_PREFIX}peer_wfe*" \
        --output ${OUT_DIR}/${OUT_PREFIX}peer_wfe

    cat ${OUT_DIR}/${OUT_PREFIX}cabundle.pem >> ${OUT_DIR}/${OUT_PREFIX}peer_wfe.pem
fi