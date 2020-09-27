#!/usr/bin/env bash
#
# Generate all trusty protobuf bindings.
# Run from repository root.
#
# This product includes software developed at CoreOS, Inc.
# (http://www.coreos.com/).
#
set -e

if ! [[ "$0" =~ scripts/build/genproto.sh ]]; then
	echo "must be run from repository root"
	exit 255
fi

if [[ $(protoc --version | cut -f2 -d' ') != "3.13.0" ]]; then
	echo "could not find protoc 3.13.0, is it installed + in PATH?"
	exit 255
fi

# directories containing protos to be built
DIRS="./api/v1/trustypb"

# exact version of packages to build
GOGO_PROTO_SHA="1adfc126b41513cc696b209667c8656ea7aac67c"
GRPC_GATEWAY_SHA="92583770e3f01b09a0d3e9bdf64321d8bebd48f2"

# disable go mod
export GO111MODULE=off
# set up self-contained GOPATH for building
export GOPATH=${PWD}/.gopath
export GOBIN=${PWD}/bin
export PATH="${GOBIN}:${PATH}"

GOGOPROTO_ROOT="${GOPATH}/src/github.com/gogo/protobuf"
GOGOPROTO_PATH="${GOGOPROTO_ROOT}:${GOGOPROTO_ROOT}/protobuf"
GRPC_GATEWAY_ROOT="${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway"

function cleanup {
  # Remove the whole fake GOPATH which can really confuse go mod.
  rm -rf "${GOPATH}"
}

#cleanup
#trap cleanup EXIT

# Ensure we have the right version of protoc-gen-gogo by building it every time.
# TODO(jonboulle): vendor this instead of `go get`ting it.
go get -u github.com/gogo/protobuf/{proto,protoc-gen-gogo,gogoproto}
go get -u golang.org/x/tools/cmd/goimports
pushd "${GOGOPROTO_ROOT}"
	git reset --hard "${GOGO_PROTO_SHA}"
	make install
popd

# generate gateway code
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
pushd "${GRPC_GATEWAY_ROOT}"
	git reset --hard "${GRPC_GATEWAY_SHA}"
	go install ./protoc-gen-grpc-gateway
popd

for dir in ${DIRS}; do
	pushd "${dir}"
   		protoc \
           --gofast_out=plugins=grpc,import_prefix=trusty_grpc/:. \
           -I=".:${GOGOPROTO_PATH}:${GRPC_GATEWAY_ROOT}/third_party/googleapis" \
           ./*.proto

		# shellcheck disable=SC1117
		sed -i.bak -E 's/trusty_grpc\/(gogoproto|github\.com|golang\.org|google\.golang\.org)/\1/g' ./*.pb.go
		# shellcheck disable=SC1117
		sed -i.bak -E 's/trusty_grpc\/(errors|fmt|io)/\1/g' ./*.pb.go
		# shellcheck disable=SC1117
		sed -i.bak -E 's/import _ \"gogoproto\"//g' ./*.pb.go
		# shellcheck disable=SC1117
		sed -i.bak -E 's/import fmt \"fmt\"//g' ./*.pb.go
		# shellcheck disable=SC1117
		sed -i.bak -E 's/import _ \"trusty_grpc\/google\/api\"//g' ./*.pb.go
		# shellcheck disable=SC1117
		sed -i.bak -E 's/import _ \"google\.golang\.org\/genproto\/googleapis\/api\/annotations\"//g' ./*.pb.go
		rm -f ./*.bak
		goimports -w ./*.pb.go
        gofmt -s -l -w ./*.pb.go
	popd
done

# remove old swagger files so it's obvious whether the files fail to generate
rm -rf Documentation/dev-guide/apispec/swagger/*json
for pb in trustypb/rpc; do
	protobase="api/v1/${pb}"
	protoc -I. \
	    -I"${GRPC_GATEWAY_ROOT}"/third_party/googleapis \
	    -I"${GOGOPROTO_PATH}" \
	    --grpc-gateway_out=logtostderr=true:. \
	    --swagger_out=logtostderr=true:./Documentation/dev-guide/apispec/swagger/. \
	    ${protobase}.proto
	# hack to move gw files around so client won't include them
	pkgpath=$(dirname "${protobase}")
	pkg=$(basename "${pkgpath}")
	gwfile="${protobase}.pb.gw.go"
	sed -i.bak -E "s/package $pkg/package gw/g" ${gwfile}
	# shellcheck disable=SC1117
	sed -i.bak -E "s/protoReq /&$pkg\./g" ${gwfile}
	sed -i.bak -E "s/, client /, client $pkg./g" ${gwfile}
	sed -i.bak -E "s/Client /, client $pkg./g" ${gwfile}
	sed -i.bak -E "s/[^(]*Client, runtime/${pkg}.&/" ${gwfile}
	sed -i.bak -E "s/New[A-Za-z]*Client/${pkg}.&/" ${gwfile}
	# darwin doesn't like newlines in sed...
	# shellcheck disable=SC1117
	sed -i.bak -E "s|import \(|& \"github.com/go-phorce/trusty/${pkgpath}\"|" ${gwfile}
	mkdir -p  "${pkgpath}"/gw/
	go fmt ${gwfile}
	mv ${gwfile} "${pkgpath}/gw/"
	rm -f ${protobase}*.bak
	swaggerName=$(basename ${pb})
	mv	Documentation/dev-guide/apispec/swagger/${protobase}.swagger.json \
		Documentation/dev-guide/apispec/swagger/"${swaggerName}".swagger.json
done
rm -rf Documentation/dev-guide/apispec/swagger/api/
