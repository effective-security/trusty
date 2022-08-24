#!/usr/bin/env bash
#
# Generate all trusty protobuf bindings.
# Run from repository root.
#
set -e

if ! [[ "$0" =~ genproto.sh ]]; then
	echo "must be run from repository root"
	exit 255
fi

if [[ $(protoc --version | cut -f2 -d' ') != "3.19.4" ]]; then
	echo "could not find protoc 3.19.4, is it installed + in PATH?"
	echo "to install:"
	echo ""
	echo "curl -L https://github.com/google/protobuf/releases/download/v3.19.4/protoc-3.19.4-linux-x86_64.zip -o /tmp/protoc.zip"
	echo "unzip /tmp/protoc.zip -d /usr/local/protoc"
	echo "export PATH=$PATH:/usr/local/protoc/bin"
	exit 255
fi

# directories containing protos to be built
DIRS="./api/v1/pb"

# exact version of packages to build
GRPC_GATEWAY_SHA="v2.4.0"

# disable go mod
export GO111MODULE=off
# set up self-contained GOPATH for building
export GOPATH=${PWD}/.gopath
export GOBIN=${PWD}/bin
export PATH="${GOBIN}:${PATH}"

GRPC_GATEWAY_ROOT="${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway"

function cleanup {
  # Remove the whole fake GOPATH which can really confuse go mod.
  rm -rf "${GOPATH}"
}

#cleanup
trap cleanup EXIT

go get github.com/golang/protobuf/{proto,ptypes,protoc-gen-go}
go get golang.org/x/tools/cmd/goimports
go get github.com/gogo/googleapis/google/api
go get github.com/ghodss/yaml
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
			-I=${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway \
			--go_out=\
Mgoogle/protobuf/empty.proto=github.com/golang/protobuf/ptypes/empty,\
Mgoogle/protobuf/struct.proto=github.com/golang/protobuf/ptypes/struct,\
Mgoogle/protobuf/duration.proto=github.com/golang/protobuf/ptypes/duration,\
Mgoogle/protobuf/timestamp.proto=github.com/golang/protobuf/ptypes/timestamp,\
Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,\
Mgoogle/api/http.proto=github.com/gogo/googleapis/google/api,\
paths=source_relative,\
plugins=grpc:. \
			./*.proto
		sed -i.bak -E 's/_ \"google\/api\"//g' ./*.pb.go
        rm -f ./*.bak
		goimports -w ./*.pb.go
		gofmt -s -l -w ./*.pb.go
	popd
done

# remove old swagger files so it's obvious whether the files fail to generate
rm -rf Documentation/dev-guide/apispec/swagger/*json
for pb in pb/status pb/cis; do
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
		-I=api/v1/pb \
		-I=${GOPATH}/src \
		-I=${GOPATH}/src/github.com/gogo/googleapis \
		-I=${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway \
		--grpc-gateway_out=\
Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/ptypes,\
Mgoogle/protobuf/empty.proto=github.com/golang/protobuf/ptypes/empty,\
Mgoogle/protobuf/struct.proto=github.com/golang/protobuf/ptypes/struct,\
Mgoogle/protobuf/duration.proto=github.com/golang/protobuf/ptypes/duration,\
Mgoogle/protobuf/timestamp.proto=github.com/golang/protobuf/ptypes/timestamp,\
Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,\
Mgoogle/api/http.proto=github.com/gogo/googleapis/google/api,\
standalone=true,\
paths=source_relative,\
allow_patch_feature=false,\
logtostderr=true:.\
		--openapiv2_out=\
Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/ptypes,\
Mgoogle/protobuf/empty.proto=github.com/golang/protobuf/ptypes/empty,\
Mgoogle/protobuf/struct.proto=github.com/golang/protobuf/ptypes/struct,\
Mgoogle/protobuf/duration.proto=github.com/golang/protobuf/ptypes/duration,\
Mgoogle/protobuf/timestamp.proto=github.com/golang/protobuf/ptypes/timestamp,\
Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,\
Mgoogle/api/http.proto=github.com/gogo/googleapis/google/api,\
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
