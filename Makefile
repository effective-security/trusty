include .project/gomod-project.mk

SHA := $(shell git rev-parse HEAD)

export COVERAGE_EXCLUSIONS="vendor|tests|main.go"
export TRUSTY_DIR=${PROJ_ROOT}
export GO111MODULE=on
BUILD_FLAGS=-mod=vendor

.PHONY: *

.SILENT:

default: help

all: clean folders tools generate hsmconfig build

#
# clean produced files
#
clean:
	echo "Running clean"
	go clean
	rm -rf \
		./bin \
		./.rpm \
		${COVPATH} \

tools:
	go install golang.org/x/tools/cmd/stringer
	go install golang.org/x/tools/cmd/gorename
	go install golang.org/x/tools/cmd/godoc
	go install golang.org/x/tools/cmd/guru
	go install golang.org/x/lint/golint
	go install github.com/go-phorce/cov-report/cmd/cov-report
	go install github.com/go-phorce/configen/cmd/configen
	go install github.com/mattn/goreman
	go install github.com/mattn/goveralls

folders:
	mkdir -p /tmp/trusty/softhsm/tokens \
		/tmp/trusty/certs \
		/tmp/trusty/logs \
		/tmp/trusty/audit \
		/tmp/trusty/certs \
		/tmp/trusty/data/etcd \
		/tmp/trusty/cluster/data_0 \
		/tmp/trusty/cluster/data_1 \
		/tmp/trusty/cluster/data_2 \
		/tmp/trusty/snapshots
	chmod -R 0700 /tmp/trusty

version:
	echo "*** building version"
	gofmt -r '"GIT_VERSION" -> "$(GIT_VERSION)"' version/current.template > version/current.go

proto:
	./scripts/build/genproto.sh

build_trusty:
	echo "*** Building trusty"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/trusty ./cmd/trusty

build_trustyctl:
	echo "*** Building trustyctl"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/trustyctl ./cmd/trustyctl

build: build_trusty build_trustyctl
	echo "Finished build"

hsmconfig:
	echo "*** Running hsmconfig"
	.project/config-softhsm.sh \
	    --tokens-dir /tmp/trusty/softhsm/tokens/ \
        --pin-file /tmp/trusty/softhsm/trusty_unittest.txt \
        --generate-pin -s trusty_unittest \
        -o /tmp/trusty/softhsm/unittest_hsm.json \
		--delete --list-slots --list-object
