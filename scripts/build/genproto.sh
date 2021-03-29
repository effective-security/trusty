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

if [[ $(protoc --version | cut -f2 -d' ') != "3.15.5" ]]; then
	echo "could not find protoc 3.15.5, is it installed + in PATH?"
	exit 255
fi

# directories containing protos to be built
DIRS="./api/v1/trustypb"

# exact version of packages to build
GOGO_PROTO_SHA="v1.3.2"
GRPC_GATEWAY_SHA="v2.3.0"

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
trap cleanup EXIT

# Ensure we have the right version of protoc-gen-gogo by building it every time.
# TODO(jonboulle): vendor this instead of `go get`ting it.
go get -u github.com/gogo/protobuf/{proto,protoc-gen-gogo,gogoproto}
go get github.com/gogo/protobuf/types
go get -u golang.org/x/tools/cmd/goimports
go get github.com/gogo/googleapis/google/api
pushd "${GOGOPROTO_ROOT}"
	git reset --hard "${GOGO_PROTO_SHA}"
	make install
popd

# generate gateway code
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-openapiv2
pushd "${GRPC_GATEWAY_ROOT}"
	git reset --hard "${GRPC_GATEWAY_SHA}"
	go install ./protoc-gen-grpc-gateway
popd

for dir in ${DIRS}; do
	pushd "${dir}"
		protoc \
			-I=. \
			-I=${GOPATH}/src \
			-I=${GOPATH}/src/github.com/gogo/googleapis \
			-I=${GOGOPROTO_PATH} \
			--gofast_out=\
Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,\
plugins=grpc:. \
			./*.proto
		sed -i.bak -E 's/_ \"google\/api\"//g' ./*.pb.go
        rm -f ./*.bak
		goimports -w ./*.pb.go
		gofmt -s -l -w ./*.pb.go
	popd
done

exit 0

# TODO: replace github.com/golang/protobuf/proto with google.golang.org/...

# remove old swagger files so it's obvious whether the files fail to generate
rm -rf Documentation/dev-guide/apispec/swagger/*json
for pb in trustypb/rpc trustypb/pkix; do
	protobase="api/v1/${pb}"
	echo "making docs and gw on: ${protobase}"
	echo "protobase=${protobase}"
	pkgpath=$(dirname "${protobase}")
	echo "pkgpath=${pkgpath}"
	pkg=$(basename "${pkgpath}")
	echo "pkg=${pkg}"
	gwfile="${protobase}.pb.gw.go"
	mkdir -p "${pkgpath}/gw/"
	protoc \
		-I=. \
		-I=api/v1/trustypb \
		-I=${GOPATH}/src \
		-I=${GOPATH}/src/github.com/gogo/googleapis \
		-I=${GOGOPROTO_PATH} \
		--grpc-gateway_out=\
Mrpc.proto=github.com/ekspand/trusty/api/v1/trustypb,\
Mapi/v1/trustypb/rpc.proto=github.com/ekspand/trusty/api/v1/trustypb,\
Mpkix.proto=github.com/ekspand/trusty/api/v1/trustypb,\
Mapi/v1/trustypb/pkix.proto=github.com/ekspand/trusty/api/v1/trustypb,\
Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,\
logtostderr=true,paths=source_relative,standalone=true:. \
		--openapiv2_out=\
Mrpc.proto=github.com/ekspand/trusty/api/v1/trustypb,\
Mapi/v1/trustypb/rpc.proto=github.com/ekspand/trusty/api/v1/trustypb,\
Mpkix.proto=github.com/ekspand/trusty/api/v1/trustypb,\
Mapi/v1/trustypb/pkix.proto=github.com/ekspand/trusty/api/v1/trustypb,\
Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,\
logtostderr=true:./Documentation/dev-guide/apispec/swagger/. \
		${protobase}.proto
	# hack to move gw files around so client won't include them
	sed -i.bak -E "s/package $pkg/package gw/g" ${gwfile}
    echo "goimports -w ${gwfile}"
	goimports -w ${gwfile}
    echo "gofmt -s -l -w ${gwfile}"
	gofmt -s -l -w ${gwfile}
    echo "mv ${gwfile} "${pkgpath}/gw/""
	mv "${gwfile}" "${pkgpath}/gw/"
	rm -rf "${gwfile}.bak"
	swaggerName=$(basename ${pb})
	mv	Documentation/dev-guide/apispec/swagger/${protobase}.swagger.json \
		Documentation/dev-guide/apispec/swagger/"${swaggerName}".swagger.json
done
rm -rf Documentation/dev-guide/apispec/swagger/api/
