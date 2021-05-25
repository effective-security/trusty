preprpm: change_log
	# this target assumes you've already done a build
	echo "Preparing files for RPM"
	rm -rf .rpm
	mkdir -p .rpm/dist
	mkdir -p .rpm/trusty/opt/trusty/bin/sql/postgres/migrations
	mkdir -p .rpm/trusty/opt/trusty/etc/prod/csr_profile
	# bin
	cp ./change_log.txt .rpm/trusty/opt/trusty/bin/
	cp bin/trusty .rpm/trusty/opt/trusty/bin/
	cp bin/trustyctl .rpm/trusty/opt/trusty/bin/
	cp bin/trusty-tool .rpm/trusty/opt/trusty/bin/
	cp ./scripts/build/*.sh .rpm/trusty/opt/trusty/bin/
	cp -R scripts/sql/ .rpm/trusty/opt/trusty/bin/
	cp -R ./Documentation .rpm/trusty/opt/trusty/
	# etc
	# cp etc/prod/*.json .rpm/trusty/opt/trusty/etc/prod/
	cp etc/prod/*.yaml .rpm/trusty/opt/trusty/etc/prod/
	cp -R etc/prod/csr_profile/ .rpm/trusty/opt/trusty/etc/prod/
	# rpm
	cp ./scripts/rpm/*.sh .rpm/
	# systemd
	mkdir -p .rpm/trusty/etc/systemd/system
	cp scripts/rpm/trusty.service .rpm/trusty/etc/systemd/system

rpm_local: preprpm
	echo "Making RPM"
	RPM_NAME=trusty
	RPM_MAINTAINER=denis@ekspand.com
	RPM_EPOCH=1
	RPM_ITER=${COMMITS_COUNT:=.el7}
	RPM_VERSION=${PROD_VERSION}
	RPM_AFTER_INSTALL=./opt/trusty/bin/postinstall.systemd.sh
	RPM_URL="https://github.com/ekspand/trusty"
	RPM_SUMMARY="Trusy service"
	./.rpm/mkrpm.sh \
		-n trusty \
		-m denis@ekspand.com \
		--epoch 1 \
		--version "${GIT_VERSION}" \
		--iteration "el7" \
		--after-install ./.rpm/mkrpm.sh \
		--url "https://github.com/ekspand/trusty" \
		--summary ${RPM_SUMMARY}

rpm_systemd: preprpm
	docker pull ekspand/docker-centos7-fpm:latest
	docker run -d -it -v ${PROJ_DIR}/.rpm:/rpm --name centos7fpm ekspand/docker-centos7-fpm
	docker exec centos7fpm ./mkrpm.sh \
		-n trusty \
		-m denis@ekspand.com \
		--epoch 1 \
		--version "${GIT_VERSION}" \
		--iteration "el7" \
		--after-install /rpm/postinstall.systemd.sh \
		--url "https://github.com/ekspand/trusty" \
		--summary "Trusy service"
	docker stop centos7fpm
	docker rm centos7fpm

rpm_docker: preprpm
	docker pull ekspand/docker-centos7-fpm:latest
	docker run -d -it -v ${PROJ_DIR}/.rpm:/rpm --name centos7fpm ekspand/docker-centos7-fpm
	docker exec centos7fpm ./mkrpm.sh \
		-n trusty \
		-m denis@ekspand.com \
		--epoch 1 \
		--version "${GIT_VERSION}-docker" \
		--iteration "el7" \
		--after-install /rpm/postinstall.sh \
		--url "https://github.com/ekspand/trusty" \
		--summary "Trusy service"
	docker stop centos7fpm
	docker rm centos7fpm
