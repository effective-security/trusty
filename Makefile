include .project/gomod-project.mk

SHA := $(shell git rev-parse HEAD)

export COVERAGE_EXCLUSIONS="vendor|tests|main.go|testsuite.go|rpc.pb.go|rpc.pb.gw.go|pkix.pb.go|pkix.pb.gw.go"
export TRUSTY_DIR=${PROJ_ROOT}
export GO111MODULE=on
BUILD_FLAGS=

.PHONY: *

.SILENT:

default: help

all: clean folders tools generate hsmconfig build gen_test_certs test

#
# clean produced files
#
clean:
	echo "Running clean"
	go clean
	rm -rf \
		./bin \
		./.rpm \
		./.gopath \
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
	gofmt -r '"GIT_VERSION" -> "$(GIT_VERSION)"' internal/version/current.template > internal/version/current.go

proto:
	./scripts/build/genproto.sh

build_trusty:
	echo "*** Building trusty"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/trusty ./cmd/trusty

build_trustyctl:
	echo "*** Building trustyctl"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/trustyctl ./cmd/trustyctl

build_tool:
	echo "*** Building tool"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/trusty-tool ./cmd/trusty-tool

build: build_trusty build_trustyctl build_tool
	echo "Finished build"

hsmconfig:
	echo "*** Running hsmconfig"
	.project/config-softhsm.sh \
		--tokens-dir /tmp/trusty/softhsm/tokens/ \
		--pin-file /tmp/trusty/softhsm/trusty_unittest.txt \
		--generate-pin -s trusty_unittest \
		-o /tmp/trusty/softhsm/unittest_hsm.json \
		--delete --list-slots --list-object

gen_test_certs:
	echo "*** Running gen_test_certs"
	echo "*** generating untrusted CAs"
	$(PROJ_ROOT)/.project/gen_test_certs.sh \
		--hsm-config /tmp/trusty/softhsm/unittest_hsm.json \
		--ca-config $(PROJ_ROOT)/etc/dev/ca-config.bootstrap.json \
		--out-dir /tmp/trusty/certs \
		--csr-dir $(PROJ_ROOT)/etc/dev/csr \
		--prefix $(PROJ_NAME)_untrusted_ \
		--root-ca /tmp/trusty/certs/trusty_untrusted_root_ca.pem \
		--root-ca-key /tmp/trusty/certs/trusty_untrusted_root_ca-key.pem \
		--root --ca1 --ca2 --bundle --peer --force
	echo "*** generating test CAs"
	rm -f /tmp/trusty/certs/$(PROJ_NAME)_dev_peer*
	$(PROJ_ROOT)/.project/gen_test_certs.sh \
		--hsm-config /tmp/trusty/softhsm/unittest_hsm.json \
		--ca-config $(PROJ_ROOT)/etc/dev/ca-config.bootstrap.json \
		--out-dir /tmp/trusty/certs \
		--csr-dir $(PROJ_ROOT)/etc/dev/csr \
		--prefix $(PROJ_NAME)_dev_ \
		--root-ca /tmp/trusty/certs/trusty_dev_root_ca.pem \
		--root-ca-key /tmp/trusty/certs/trusty_dev_root_ca-key.pem \
		--root --ca1 --ca2  --bundle --peer --force

start-sql:
	# Container state will be true (it's already running), false (exists but stopped), or missing (does not exist).
	# Annoyingly, when there is no such container and Docker returns an error, it also writes a blank line to stdout.
	# Hence the sed to trim whitespace.
	CONTAINER_STATE=$$(echo $$(docker inspect -f "{{.State.Running}}" trusty-unittest-postgres 2>/dev/null || echo "missing") | sed -e 's/^[ \t]*//'); \
	if [ "$$CONTAINER_STATE" = "missing" ]; then \
		docker pull ekspand/docker-centos7-postgres:latest && \
		docker run --network host \
			-d -p 5432:5432 \
			-e 'POSTGRES_USER=postgres' -e 'POSTGRES_PASSWORD=postgres' \
			-v ${PROJ_ROOT}/scripts/sql/postgres:/scripts/pgsql \
			--name trusty-unittest-postgres \
			ekspand/docker-centos7-postgres:latest /start_postgres.sh && \
		sleep 9; \
	elif [ "$$CONTAINER_STATE" = "false" ]; then docker start trusty-unittest-postgres && sleep 9; fi;
	docker exec -e 'PGPASSWORD=postgres' trusty-unittest-postgres psql -h 127.0.0.1 -p 5432 -U postgres -a -f /scripts/pgsql/create-trustrydb.sql
	docker exec -e 'PGPASSWORD=postgres' trusty-unittest-postgres psql -h 127.0.0.1 -p 5432 -U postgres -a -f /scripts/pgsql/list-trustrydb.sql
	echo "host=127.0.0.1 port=5432 user=postgres password=postgres sslmode=disable dbname=trustydb" > etc/dev/sql-conn.txt

drop-sql:
	docker pull ekspand/docker-centos7-postgres:latest
	docker exec -e 'PGPASSWORD=postgres' trusty-unittest-postgres psql -h 127.0.0.1 -p 5432 -U postgres -a -f /scripts/pgsql/drop-trustydb.sql
	docker exec -e 'PGPASSWORD=postgres' trusty-unittest-postgres psql -h 127.0.0.1 -p 5432 -U postgres -lqt

coveralls-github:
	echo "Running coveralls"
	goveralls -v -coverprofile=coverage.out -service=github -package ./...
