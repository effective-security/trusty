include .project/gomod-project.mk

SHA := $(shell git rev-parse HEAD)

export AWS_ACCESS_KEY_ID=notusedbyemulator
export AWS_SECRET_ACCESS_KEY=notusedbyemulator
export AWS_DEFAULT_REGION=us-west-2
export TRUSTY_JWT_SEED=testseed 

export COVERAGE_EXCLUSIONS="vendor|tests|api/v1/pb/gw|main.go|testsuite.go|mocks.go|.pb.go|.pb.gw.go"
export TRUSTY_DIR=${PROJ_ROOT}
export GO111MODULE=on
BUILD_FLAGS=

.PHONY: *

.SILENT:

default: help

all: clean folders tools generate hsmconfig build gen_test_certs gen_martini_certs start-local-kms start-sql test

#
# clean produced files
#
clean:
	echo "Running clean"
	go clean
	rm -rf \
		./bin \
		./.gopath \
		${COVPATH} \

tools:
	go install golang.org/x/tools/cmd/stringer
	go install golang.org/x/tools/cmd/gorename
	go install golang.org/x/tools/cmd/godoc
	go install golang.org/x/tools/cmd/guru
	go install golang.org/x/lint/golint
	go install github.com/go-phorce/cov-report/cmd/cov-report
	go install github.com/mattn/goreman
	go install github.com/mattn/goveralls
	go install golang.org/x/tools/cmd/goimports
	#go install sigs.k8s.io/controller-tools/cmd/controller-gen
	# go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	# go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

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
	./genproto.sh

build_trusty:
	echo "*** Building trusty"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/trusty ./cmd/trusty

build_trustyctl:
	echo "*** Building trustyctl"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/trustyctl ./cmd/trustyctl

build_tool:
	echo "*** Building tool"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/trusty-tool ./cmd/trusty-tool

build_kube:
	echo "*** Building kubeca"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/kubeca ./cmd/kubeca
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/kubecertinit ./cmd/kubecertinit

build_martinictl:
	echo "*** Building martinictl"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/martinictl ./cmd/martinictl

build: build_trusty build_trustyctl build_tool build_kube build_martinictl

change_log:
	echo "Recent changes:" > ./change_log.txt
	git log -n 20 --pretty=oneline --abbrev-commit >> ./change_log.txt

commit_version:
	git add .; git commit -m "Updated version"

hsmconfig:
	echo "*** Running hsmconfig"
	./scripts/build/config-softhsm.sh \
		--tokens-dir /tmp/trusty/softhsm/tokens/ \
		--pin-file /tmp/trusty/softhsm/trusty_unittest.txt \
		--generate-pin -s trusty_unittest \
		-o /tmp/trusty/softhsm/unittest_hsm.json \
		--delete --list-slots --list-object

gen_test_certs:
	echo "*** Running gen_test_certs"
	echo "*** generating untrusted CAs"
	./scripts/build/gen_test_certs.sh \
		--hsm-config /tmp/trusty/softhsm/unittest_hsm.json \
		--ca-config $(PROJ_ROOT)/etc/dev/ca-config.bootstrap.yaml \
		--out-dir /tmp/trusty/certs \
		--csr-dir $(PROJ_ROOT)/etc/dev/csr_profile \
		--prefix $(PROJ_NAME)_untrusted_ \
		--key-label unittest_ \
		--root-ca /tmp/trusty/certs/trusty_untrusted_root_ca.pem \
		--root-ca-key /tmp/trusty/certs/trusty_untrusted_root_ca-key.pem \
		--root --ca1 --ca2 --bundle --client --force
	echo "*** generating test CAs"
	tar -xzvf $(PROJ_ROOT)/etc/dev/roots/trusty_dev_root_ca-key.pem.tar.gz -C $(PROJ_ROOT)/etc/dev/roots/
	./scripts/build/gen_test_certs.sh \
		--hsm-config /tmp/trusty/softhsm/unittest_hsm.json \
		--ca-config $(PROJ_ROOT)/etc/dev/ca-config.bootstrap.yaml \
		--out-dir /tmp/trusty/certs \
		--csr-dir $(PROJ_ROOT)/etc/dev/csr_profile \
		--prefix $(PROJ_NAME)_dev_ \
		--key-label test_ \
		--root-ca $(PROJ_ROOT)/etc/dev/roots/trusty_dev_root_ca.pem \
		--root-ca-key $(PROJ_ROOT)/etc/dev/roots/trusty_dev_root_ca-key.pem \
		--ca1 --ca2  --bundle --peer --client --force
	cp $(PROJ_ROOT)/etc/dev/roots/trusty_dev_root_ca.pem /tmp/trusty/certs/trusty_dev_root_ca.pem

gen_martini_certs:
	echo "*** Running gen_martini_certs"
	./scripts/build/gen_martini_certs.sh \
		--hsm-config /tmp/trusty/softhsm/unittest_hsm.json \
		--ca-config $(PROJ_ROOT)/etc/dev/ca-config.bootstrap.yaml \
		--out-dir /tmp/trusty/certs \
		--csr-dir $(PROJ_ROOT)/etc/dev/csr_profile \
		--prefix martini_ \
		--root-ca /tmp/trusty/certs/martini_root_ca.pem \
		--root-ca-key /tmp/trusty/certs/martini_root_ca-key.pem \
		--root --ca1 --bundle --signer --force

start-local-kms:
	# Container state will be true (it's already running), false (exists but stopped), or missing (does not exist).
	# Annoyingly, when there is no such container and Docker returns an error, it also writes a blank line to stdout.
	# Hence the sed to trim whitespace.
	LKMS_CONTAINER_STATE=$$(echo $$(docker inspect -f "{{.State.Running}}" trusty-unittest-local-kms 2>/dev/null || echo "missing") | sed -e 's/^[ \t]*//'); \
	if [ "$$LKMS_CONTAINER_STATE" = "missing" ]; then \
		docker pull nsmithuk/local-kms && \
		docker run \
			-d -e 'PORT=4599' \
			-p 4599:4599 \
			--name trusty-unittest-local-kms \
			nsmithuk/local-kms && \
			sleep 1; \
	elif [ "$$LKMS_CONTAINER_STATE" = "false" ]; then docker start trusty-unittest-local-kms && sleep 1; fi;

start-sql:
	# Container state will be true (it's already running), false (exists but stopped), or missing (does not exist).
	# Annoyingly, when there is no such container and Docker returns an error, it also writes a blank line to stdout.
	# Hence the sed to trim whitespace.
	CONTAINER_STATE=$$(echo $$(docker inspect -f "{{.State.Running}}" trusty-unittest-postgres 2>/dev/null || echo "missing") | sed -e 's/^[ \t]*//'); \
	if [ "$$CONTAINER_STATE" = "missing" ]; then \
		docker pull ekspand/docker-centos7-postgres:latest && \
		docker run \
			-d -p 5432:5432 \
			-e 'POSTGRES_USER=postgres' -e 'POSTGRES_PASSWORD=postgres' \
			-v ${PROJ_ROOT}/sql:/trusty_sql \
			--name trusty-unittest-postgres \
			ekspand/docker-centos7-postgres:latest /start_postgres.sh && \
		sleep 9; \
	elif [ "$$CONTAINER_STATE" = "false" ]; then docker start trusty-unittest-postgres && sleep 9; fi;
	docker exec -e 'PGPASSWORD=postgres' trusty-unittest-postgres psql -h 127.0.0.1 -p 5432 -U postgres -a -f /trusty_sql/cadb/create.sql
	docker exec -e 'PGPASSWORD=postgres' trusty-unittest-postgres psql -h 127.0.0.1 -p 5432 -U postgres -a -f /trusty_sql/orgsdb/create.sql
	docker exec -e 'PGPASSWORD=postgres' trusty-unittest-postgres psql -h 127.0.0.1 -p 5432 -U postgres -lqt
	echo "host=127.0.0.1 port=5432 user=postgres password=postgres sslmode=disable dbname=cadb" > etc/dev/sql-conn-cadb.txt
	echo "host=127.0.0.1 port=5432 user=postgres password=postgres sslmode=disable dbname=orgsdb" > etc/dev/sql-conn-orgsdb.txt

drop-sql:
	docker exec -e 'PGPASSWORD=postgres' trusty-unittest-postgres psql -h 127.0.0.1 -p 5432 -U postgres -a -f /trusty_sql/cadb/drop.sql
	docker exec -e 'PGPASSWORD=postgres' trusty-unittest-postgres psql -h 127.0.0.1 -p 5432 -U postgres -a -f /trusty_sql/orgsdb/drop.sql
	docker exec -e 'PGPASSWORD=postgres' trusty-unittest-postgres psql -h 127.0.0.1 -p 5432 -U postgres -lqt

coveralls-github:
	echo "Running coveralls"
	goveralls -v -coverprofile=coverage.out -service=github -package ./...

docker:
	docker build --no-cache -f Dockerfile -t ekspand/trusty:master .

docker-kubeca:	
	docker build --no-cache -f Dockerfile.kubeca -t ekspand/kubeca:master .
	docker build --no-cache -f Dockerfile.kubecertinit -t ekspand/kubecertinit:master .

docker-compose:
	docker-compose -f docker-compose.dev.yml up --abort-on-container-exit

docker-push: docker
	[ ! -z ${DOCKER_PASSWORD} ] && echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin || echo "skipping docker login"
	docker push ekspand/trusty:latest
	#[ ! -z ${DOCKER_NUMBER} ] && docker push ekspand/trusty:${DOCKER_NUMBER} || echo "skipping docker version, pushing latest only"

docker-citest:
	cd ./scripts/integration && ./setup.sh

start-swagger:
	CONTAINER_STATE=$$(echo $$(docker inspect -f "{{.State.Running}}" trusty-swagger 2>/dev/null || echo "missing") | sed -e 's/^[ \t]*//'); \
	if [ "$$CONTAINER_STATE" = "missing" ]; then \
		docker pull swaggerapi/swagger-ui && \
		docker run \
		-d -p 8080:8080 \
		-e URLS="[ { url: \"http://localhost:7880/v1/swagger/status\", name: \"StatusService\" }, { url: \"http://localhost:7880/v1/swagger/cis\", name: \"CIService\" }, { url: \"https://localhost:7892/v1/swagger/ca\", name: \"CAService\" } ]" \
		-v ${PROJ_DIR}/Documentation/dev-guide/apispec/swagger:/swagger \
		--name trusty-swagger \
		swaggerapi/swagger-ui ; \
	elif [ "$$CONTAINER_STATE" = "false" ]; then docker start trusty-swagger; fi;

start-prometheus:
	CONTAINER_STATE=$$(echo $$(docker inspect -f "{{.State.Running}}" trusty-prometheus 2>/dev/null || echo "missing") | sed -e 's/^[ \t]*//'); \
	if [ "$$CONTAINER_STATE" = "missing" ]; then \
		docker pull prom/prometheus && \
		docker run \
			-d  -p 9090:9090 \
			-v ${PROJ_DIR}/etc/dev/prometheus.yml:/etc/prometheus/prometheus.yml \
			--name trusty-prometheus \
			prom/prometheus ;\
	elif [ "$$CONTAINER_STATE" = "false" ]; then docker start trusty-prometheus; fi;

start-grafana:
	CONTAINER_STATE=$$(echo $$(docker inspect -f "{{.State.Running}}" trusty-grafana 2>/dev/null || echo "missing") | sed -e 's/^[ \t]*//'); \
	if [ "$$CONTAINER_STATE" = "missing" ]; then \
		docker pull grafana/grafana && \
		docker run \
		-d -p 3000:3000 \
		--name trusty-grafana \
		grafana/grafana ;\
	elif [ "$$CONTAINER_STATE" = "false" ]; then docker start trusty-grafana; fi;

start-pgadmin:
	CONTAINER_STATE=$$(echo $$(docker inspect -f "{{.State.Running}}" trusty-pgadmin 2>/dev/null || echo "missing") | sed -e 's/^[ \t]*//'); \
	if [ "$$CONTAINER_STATE" = "missing" ]; then \
		docker pull dpage/pgadmin4 && \
		docker run \
			-e 'PGADMIN_DEFAULT_EMAIL=admin@trusty.com' \
			-e 'PGADMIN_DEFAULT_PASSWORD=trusty' \
			-d -p 8888:80 \
			--name trusty-pgadmin \
			dpage/pgadmin4 ;\
	elif [ "$$CONTAINER_STATE" = "false" ]; then docker start trusty-pgadmin; fi;
