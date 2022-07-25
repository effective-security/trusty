include .project/gomod-project.mk

SHA := $(shell git rev-parse HEAD)

export AWS_ACCESS_KEY_ID=notusedbyemulator
export AWS_SECRET_ACCESS_KEY=notusedbyemulator
export AWS_DEFAULT_REGION=us-west-2
export TRUSTY_JWT_SEED=testseed

export GOPRIVATE=github.com/effective-security,github.com/go-phorce
export COVERAGE_EXCLUSIONS="vendor|tests|api/v1/pb/gw|main.go|testsuite.go|mocks.go|.pb.go|.pb.gw.go"
export TRUSTY_DIR=${PROJ_ROOT}
export GO111MODULE=on
BUILD_FLAGS=

.PHONY: *

.SILENT:

default: help

all: clean folders tools generate start-local-deps build gen_test_certs gen_shaken_certs test change_log

#
# clean produced files
#
clean:
	echo "Running clean"
	go clean
	rm -rf \
		/tmp/trusty \
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
	go install github.com/effective-security/xpki/cmd/hsm-tool
	go install github.com/effective-security/xpki/cmd/xpki-tool

folders:
	mkdir -p /tmp/trusty/softhsm/tokens \
		/tmp/trusty/certs \
		/tmp/trusty/logs \
		/tmp/trusty/certs
	chmod -R 0700 /tmp/trusty

version:
	echo "*** building version $(GIT_VERSION)"
	gofmt -r '"GIT_VERSION" -> "$(GIT_VERSION)"' internal/version/current.template > internal/version/current.go

proto:
	./genproto.sh

build_trusty:
	echo "*** Building trusty"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/trusty ./cmd/trusty

build_trustyctl:
	echo "*** Building trustyctl"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/trustyctl ./cmd/trustyctl

build: build_trusty build_trustyctl

change_log:
	echo "Recent changes" > ./change_log.txt
	echo "Build Version: $(GIT_VERSION)" >> ./change_log.txt
	echo "Commit: $(GIT_HASH)" >> ./change_log.txt
	echo "==================================" >> ./change_log.txt
	git log -n 20 --pretty=oneline --abbrev-commit >> ./change_log.txt

commit_version:
	git add .; git commit -m "Updated version"

gen_test_certs:
	echo "*** Running gen_test_certs"
	echo "*** generating test CAs"
	tar -xzvf $(PROJ_ROOT)/etc/dev/roots/trusty_root_ca.key.tar.gz -C $(PROJ_ROOT)/etc/dev/roots/
	./scripts/build/gen_certs.sh \
		--hsm-config $(PROJ_ROOT)/etc/dev/kms/aws-dev-kms.yaml \
		--ca-config $(PROJ_ROOT)/etc/dev/ca-config.bootstrap.yaml \
		--output-dir /tmp/trusty/certs \
		--csr-dir $(PROJ_ROOT)/etc/dev/csr_profile \
		--csr-prefix trusty_ \
		--out-prefix trusty_ \
		--key-label test_ \
		--root-ca $(PROJ_ROOT)/etc/dev/roots/trusty_root_ca.pem \
		--root-ca-key $(PROJ_ROOT)/etc/dev/roots/trusty_root_ca.key \
		--ca1 --ca2  --peer --bundle --force
	cp $(PROJ_ROOT)/etc/dev/roots/trusty_root_ca.pem /tmp/trusty/certs/trusty_root_ca.pem

gen_shaken_certs:
	./scripts/build/gen_shaken_certs.sh \
		--hsm-config $(PROJ_ROOT)/etc/dev/kms/aws-dev-kms-shaken.yaml \
		--crypto $(PROJ_ROOT)/etc/dev/kms/aws-dev-kms.yaml \
		--ca-config $(PROJ_ROOT)/etc/dev/ca-config.bootstrap.yaml \
		--out-dir /tmp/trusty/certs \
		--csr-dir $(PROJ_ROOT)/etc/dev/csr_profile \
		--csr-prefix shaken_ \
		--out-prefix shaken_ \
		--root-ca /tmp/trusty/certs/shaken_root_ca.pem \
		--root-ca-key /tmp/trusty/certs/shaken_root_ca.key \
		--root --ca --l1_ca --bundle --force

start-local-deps:
	echo "*** starting local-kms"
	docker-compose -f docker-compose.trusty.yml -p trusty up -d --force-recreate --remove-orphans
	sleep 3
	docker exec -e 'PGPASSWORD=postgres' trusty_sql_1 psql -h localhost -p 15432 -U postgres -a -f /trusty_sql/cadb/create.sql
	docker exec -e 'PGPASSWORD=postgres' trusty_sql_1 psql -h localhost -p 15432 -U postgres -lqt
	echo "host=localhost port=15432 user=postgres password=postgres sslmode=disable dbname=cadb" > etc/dev/sql-conn-cadb.txt

drop-sql:
	docker exec -e 'PGPASSWORD=postgres' trusty_sql_1 psql -h localhost -p 15432 -U postgres -a -f /trusty_sql/cadb/drop.sql
	docker exec -e 'PGPASSWORD=postgres' trusty_sql_1 psql -h localhost -p 15432 -U postgres -lqt

coveralls-github:
	echo "Running coveralls"
	goveralls -v -coverprofile=coverage.out -service=github -package ./...

docker: change_log
	docker build --no-cache -f Dockerfile -t effectivesecurity/trusty:main .

docker-compose:
	docker-compose -f docker-compose.dev.yml up --abort-on-container-exit

docker-push: docker
	[ ! -z ${DOCKER_PASSWORD} ] && echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin || echo "skipping docker login"
	docker push effectivesecurity/trusty:main
	#[ ! -z ${DOCKER_NUMBER} ] && docker push effectivesecurity/trusty:${DOCKER_NUMBER} || echo "skipping docker version, pushing latest only"

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
