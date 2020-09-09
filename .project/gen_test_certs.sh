#!/bin/bash

#
# gen_test_certs.sh
#   --out-dir {dir}     - specifies output folder
#   --csr-dir {dir}     - specifies folder with CSR templates
#   --prefix {prefix}   - specifies prefix for files, by default: ${PREFIX}
#   --ca-config         - specifies CA configurationn file
#   --root-ca {cert}    - specifies root CA certificate
#   --root-ca-key {key} - specifies root CA key
#   --root              - specifies if Root CA certificate and key should be generated
#   --ca1               - specifies if Level 1 CA certificate and key should be generated
#   --ca2               - specifies if Level 2 CA certificate and key should be generated
#   --server            - specifies if server TLS certificate and key should be generated
#   --client            - specifies if client certificate and key should be generated
#   --peer              - specifies if peer certificate and key should be generated
#   --bundle            - specifies if Int CA Bundle should be created
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
    *)
    echo "invalid flag $key: use --help to see the option"
    exit 1
esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

[ -z "$OUT_DIR" ] &&  echo "Specify --out-dir" && exit 1
[ -z "$CSR_DIR" ] &&  echo "Specify --csr-dir" && exit 1
[ -z "$CA_CONFIG" ] && echo "Specify --ca-config" && exit 1
[ -z "$PREFIX" ] && PREFIX=test_
[ -z "$ROOT_CA_CERT" ] && ROOT_CA_CERT=${OUT_DIR}/${PREFIX}root_ca.pem
[ -z "$ROOT_CA_KEY" ] && ROOT_CA_KEY=${OUT_DIR}/${PREFIX}root_ca-key.pem

HOSTNAME=`hostname`

echo "OUT_DIR      = ${OUT_DIR}"
echo "CSR_DIR      = ${CSR_DIR}"
echo "CA_CONFIG    = ${CA_CONFIG}"
echo "PREFIX       = ${PREFIX}"
echo "BUNDLE       = ${BUNDLE}"
echo "FORCE        = ${FORCE}"
echo "ROOT_CA_CERT = $ROOT_CA_CERT"
echo "ROOT_CA_KEY  = $ROOT_CA_KEY"

if [[ "$ROOTCA" == "YES" && ("$FORCE" == "YES" || ! -f ${ROOT_CA_KEY}) ]]; then echo "*** generating ${ROOT_CA_CERT/.pem/''}"
    cfssl genkey -initca -config=${CA_CONFIG} ${CSR_DIR}/${PREFIX}root_ca.json | cfssljson -bare ${ROOT_CA_CERT/.pem/''}
fi

if [[ "$CA1" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}issuer1_ca-key.pem) ]]; then
    echo "*** generating CA1 cert"
    cfssl genkey -initca \
        -config=${CA_CONFIG} \
        ${CSR_DIR}/${PREFIX}issuer1_ca.json | cfssljson -bare ${OUT_DIR}/${PREFIX}issuer1_ca

    cfssl sign \
        -config=${CA_CONFIG} \
        -profile=L1_CA \
        -ca ${ROOT_CA_CERT} \
        -ca-key ${ROOT_CA_KEY} \
        -csr ${OUT_DIR}/${PREFIX}issuer1_ca.csr | cfssljson -bare ${OUT_DIR}/${PREFIX}issuer1_ca
fi

if [[ "$CA2" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem) ]]; then
    echo "*** generating CA2 cert"
    cfssl genkey -initca -config=${CA_CONFIG} ${CSR_DIR}/${PREFIX}issuer2_ca.json | cfssljson -bare ${OUT_DIR}/${PREFIX}issuer2_ca

    cfssl sign \
        -config=${CA_CONFIG} \
        -profile=L2_CA \
        -ca ${OUT_DIR}/${PREFIX}issuer1_ca.pem \
        -ca-key ${OUT_DIR}/${PREFIX}issuer1_ca-key.pem \
        -csr ${OUT_DIR}/${PREFIX}issuer2_ca.csr | cfssljson -bare ${OUT_DIR}/${PREFIX}issuer2_ca
fi

if [[ "$BUNDLE" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}cabundle.pem) ]]; then
    echo "*** CA bundle"
    cat ${OUT_DIR}/${PREFIX}issuer2_ca.pem > ${OUT_DIR}/${PREFIX}cabundle.pem
    cat ${OUT_DIR}/${PREFIX}issuer1_ca.pem >> ${OUT_DIR}/${PREFIX}cabundle.pem
fi

if [[ "$ADMIN" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}admin-key.pem) ]]; then
    echo "*** generating admin cert"
    cfssl gencert \
        -config=${CA_CONFIG} \
        -profile=client \
        -ca ${OUT_DIR}/${PREFIX}issuer2_ca.pem \
        -ca-key ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem \
        ${CSR_DIR}/${PREFIX}admin.json | cfssljson -bare ${OUT_DIR}/${PREFIX}admin
        cat ${OUT_DIR}/${PREFIX}cabundle.pem >> ${OUT_DIR}/${PREFIX}admin.pem
fi

if [[ "$SERVER" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}server-key.pem) ]]; then
    echo "*** generating server cert"
    cfssl gencert \
        -config=${CA_CONFIG} \
        -profile=server \
        -ca ${OUT_DIR}/${PREFIX}issuer2_ca.pem \
        -ca-key ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem \
        -hostname=localhost,127.0.0.1,${HOSTNAME} \
        ${CSR_DIR}/${PREFIX}server.json | cfssljson -bare ${OUT_DIR}/${PREFIX}server
        cat ${OUT_DIR}/${PREFIX}cabundle.pem >> ${OUT_DIR}/${PREFIX}server.pem
fi

if [[ "$CLIENT" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}client-key.pem) ]]; then
    echo "*** generating client cert"
    cfssl gencert \
        -config=${CA_CONFIG} \
        -profile=client \
        -ca ${OUT_DIR}/${PREFIX}issuer2_ca.pem \
        -ca-key ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem \
        ${CSR_DIR}/${PREFIX}client.json | cfssljson -bare ${OUT_DIR}/${PREFIX}client
        cat ${OUT_DIR}/${PREFIX}cabundle.pem >> ${OUT_DIR}/${PREFIX}client.pem
fi

if [[ "$PEERS" == "YES" && ("$FORCE" == "YES" || ! -f ${OUT_DIR}/${PREFIX}peer-key.pem) ]]; then
    echo "*** generating peer cert"
    cfssl gencert \
        -config=${CA_CONFIG} \
        -profile=peer \
        -ca ${OUT_DIR}/${PREFIX}issuer2_ca.pem \
        -ca-key ${OUT_DIR}/${PREFIX}issuer2_ca-key.pem \
        -hostname=localhost,127.0.0.1,${HOSTNAME} \
        ${CSR_DIR}/${PREFIX}peer.json | cfssljson -bare ${OUT_DIR}/${PREFIX}peer
        cat ${OUT_DIR}/${PREFIX}cabundle.pem >> ${OUT_DIR}/${PREFIX}peer.pem
fi
