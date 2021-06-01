#!/bin/bash

#
# gen_test_certs.sh
#   --out-dir {dir}     - specifies output folder
#   --csr-dir {dir}     - specifies folder with CSR templates
#   --prefix {prefix}   - specifies prefix for files, by default: ${PREFIX}
#   --hsm-confg         - specifies HSM provider file
#   --ca-config         - specifies CA configuration file
#   --root-ca {cert}    - specifies root CA certificate
#   --root-ca-key {key} - specifies root CA key
#   --root              - specifies if Root CA certificate and key should be generated
#   --ca1               - specifies if Level 1 CA certificate and key should be generated
#   --ca2               - specifies if Level 2 CA certificate and key should be generated
#   --server            - specifies if server TLS certificate and key should be generated
#   --client            - specifies if client certificate and key should be generated
#   --peer              - specifies if peer certificate and key should be generated
#   --bundle            - specifies if Int CA Bundle should be created
#   --san               - specifies SAN for server and peer certs
#   --force             - specifies to force issuing the cert even if it exists
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
    -o|--csr-dir)
    CSR_DIR="$2"
    shift # past argument
    shift # past value
    ;;
    -p|--prefix)
    PREFIX="$2"
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
[ -z "$PREFIX" ] && PREFIX=test_
[ -z "$ROOT_CA_CERT" ] && ROOT_CA_CERT=${OUT_DIR}/${PREFIX}root_ca.pem
[ -z "$ROOT_CA_KEY" ] && ROOT_CA_KEY=${OUT_DIR}/${PREFIX}root_ca-key.pem
[ -z "$SAN" ] && SAN=127.0.0.1

HOSTNAME=`hostname`

echo "OUT_DIR      = ${OUT_DIR}"
echo "CSR_DIR      = ${CSR_DIR}"
echo "CA_CONFIG    = ${CA_CONFIG}"
echo "HSM_CONFIG   = ${HSM_CONFIG}"
echo "PREFIX       = ${PREFIX}"
echo "BUNDLE       = ${BUNDLE}"
echo "FORCE        = ${FORCE}"
echo "SAN          = ${SAN}"
echo "ROOT_CA_CERT = $ROOT_CA_CERT"
echo "ROOT_CA_KEY  = $ROOT_CA_KEY"

if [[ "$ROOTCA" == "YES" && ("$FORCE" == "YES" || ! -f ${ROOT_CA_KEY}) ]]; then echo "*** generating ${ROOT_CA_CERT/.pem/''}"
    trusty-tool \
        --hsm-cfg=inmem \
        csr gencert --plain-key --self-sign \
        --ca-config=${CA_CONFIG} \
        --profile=ROOT \
        --csr-profile ${CSR_DIR}/${PREFIX}root_ca.json \
        --key-label="${PREFIX}root_ca*" \
        --out ${ROOT_CA_CERT/.pem/''}
fi

if [[ "$CA1" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}issuer1_ca-key.pem) ]]; then
    echo "*** generating CA1 cert"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} \
        csr gencert \
        --ca-config=${CA_CONFIG} \
        --profile=L1_CA \
        --csr-profile ${CSR_DIR}/${PREFIX}issuer1_ca.json \
        --key-label="${PREFIX}issuer1_ca*" \
        --ca-cert ${ROOT_CA_CERT} \
        --ca-key ${ROOT_CA_KEY} \
        --out ${OUT_DIR}/${PREFIX}issuer1_ca
fi

if [[ "$CA2" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem) ]]; then
    echo "*** generating CA2 cert"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} \
        csr gencert \
        --ca-config=${CA_CONFIG} \
        --profile=L2_CA \
        --csr-profile ${CSR_DIR}/${PREFIX}issuer2_ca.json \
        --key-label="${PREFIX}issuer2_ca*" \
        --ca-cert ${OUT_DIR}/${PREFIX}issuer1_ca.pem \
        --ca-key ${OUT_DIR}/${PREFIX}issuer1_ca-key.pem \
        --out ${OUT_DIR}/${PREFIX}issuer2_ca
fi

if [[ "$BUNDLE" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}cabundle.pem) ]]; then
    echo "*** CA bundle"
    cat ${OUT_DIR}/${PREFIX}issuer2_ca.pem > ${OUT_DIR}/${PREFIX}cabundle.pem
    cat ${OUT_DIR}/${PREFIX}issuer1_ca.pem >> ${OUT_DIR}/${PREFIX}cabundle.pem
fi

if [[ "$ADMIN" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}admin-key.pem) ]]; then
    echo "*** generating admin cert"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} \
        csr gencert --plain-key \
        --ca-config=${CA_CONFIG} \
        --profile=client \
        --ca-cert ${OUT_DIR}/${PREFIX}issuer2_ca.pem \
        --ca-key ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem \
        --csr-profile ${CSR_DIR}/${PREFIX}admin.json \
        --key-label="${PREFIX}admin*" \
        --out ${OUT_DIR}/${PREFIX}admin

        cat ${OUT_DIR}/${PREFIX}cabundle.pem >> ${OUT_DIR}/${PREFIX}admin.pem
fi

if [[ "$SERVER" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}server-key.pem) ]]; then
    echo "*** generating server cert"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} \
        csr gencert --plain-key \
        --ca-config=${CA_CONFIG} \
        --profile=server \
        --ca-cert ${OUT_DIR}/${PREFIX}issuer2_ca.pem \
        --ca-key ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem \
        --csr-profile ${CSR_DIR}/${PREFIX}server.json \
        --SAN=localhost,${SAN},${HOSTNAME} \
        --key-label="${PREFIX}server*" \
        --out ${OUT_DIR}/${PREFIX}server

        cat ${OUT_DIR}/${PREFIX}cabundle.pem >> ${OUT_DIR}/${PREFIX}server.pem
fi

if [[ "$CLIENT" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}client-key.pem) ]]; then
    echo "*** generating client cert"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} \
        csr gencert --plain-key \
        --ca-config=${CA_CONFIG} \
        --profile=client \
        --ca-cert ${OUT_DIR}/${PREFIX}issuer2_ca.pem \
        --ca-key ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem \
        --csr-profile ${CSR_DIR}/${PREFIX}client.json \
        --SAN=spifee://trusty/all \
        --key-label="${PREFIX}client*" \
        --out ${OUT_DIR}/${PREFIX}client

    cat ${OUT_DIR}/${PREFIX}cabundle.pem >> ${OUT_DIR}/${PREFIX}client.pem
fi

if [[ "$PEERS" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}peer_ca-key.pem) ]]; then
    echo "*** generating peer_ca cert"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} \
        csr gencert --plain-key \
        --ca-config=${CA_CONFIG} \
        --profile=peer \
        --ca-cert ${OUT_DIR}/${PREFIX}issuer2_ca.pem \
        --ca-key ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem \
        --csr-profile ${CSR_DIR}/${PREFIX}peer_ca.json \
        --SAN=localhost,${SAN},${HOSTNAME},spifee://trusty/ca \
        --key-label="${PREFIX}peer_ca*" \
        --out ${OUT_DIR}/${PREFIX}peer_ca

    cat ${OUT_DIR}/${PREFIX}cabundle.pem >> ${OUT_DIR}/${PREFIX}peer_ca.pem
fi

if [[ "$PEERS" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}peer_ra-key.pem) ]]; then
    echo "*** generating peer_ra cert"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} \
        csr gencert --plain-key \
        --ca-config=${CA_CONFIG} \
        --profile=peer \
        --ca-cert ${OUT_DIR}/${PREFIX}issuer2_ca.pem \
        --ca-key ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem \
        --csr-profile ${CSR_DIR}/${PREFIX}peer_ra.json \
        --SAN=localhost,${SAN},${HOSTNAME},spifee://trusty/ra \
        --key-label="${PREFIX}peer_ra*" \
        --out ${OUT_DIR}/${PREFIX}peer_ra

    cat ${OUT_DIR}/${PREFIX}cabundle.pem >> ${OUT_DIR}/${PREFIX}peer_ra.pem
fi

if [[ "$PEERS" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}peer_cis-key.pem) ]]; then
    echo "*** generating peer_cis cert"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} \
        csr gencert --plain-key \
        --ca-config=${CA_CONFIG} \
        --profile=peer \
        --ca-cert ${OUT_DIR}/${PREFIX}issuer2_ca.pem \
        --ca-key ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem \
        --csr-profile ${CSR_DIR}/${PREFIX}peer_cis.json \
        --SAN=localhost,${SAN},${HOSTNAME},spifee://trusty/cis \
        --key-label="${PREFIX}peer_cis*" \
        --out ${OUT_DIR}/${PREFIX}peer_cis

    cat ${OUT_DIR}/${PREFIX}cabundle.pem >> ${OUT_DIR}/${PREFIX}peer_cis.pem
fi

if [[ "$PEERS" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}peer_wfe-key.pem) ]]; then
    echo "*** generating peer_wfe cert"
    trusty-tool \
        --hsm-cfg=${HSM_CONFIG} \
        csr gencert --plain-key \
        --ca-config=${CA_CONFIG} \
        --profile=peer \
        --ca-cert ${OUT_DIR}/${PREFIX}issuer2_ca.pem \
        --ca-key ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem \
        --csr-profile ${CSR_DIR}/${PREFIX}peer_wfe.json \
        --SAN=localhost,${SAN},${HOSTNAME},spifee://trusty/wfe \
        --key-label="${PREFIX}peer_wfe*" \
        --out ${OUT_DIR}/${PREFIX}peer_wfe

    cat ${OUT_DIR}/${PREFIX}cabundle.pem >> ${OUT_DIR}/${PREFIX}peer_wfe.pem
fi