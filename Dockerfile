FROM ghcr.io/ekspand/trusty-docker-base:latest
LABEL org.opencontainers.image.source https://github.com/ekspand/trusty

ENV TRUSTY_DIR=/opt/trusty
ENV PATH=$PATH:/opt/trusty/bin
RUN mkdir -p $TRUSTY_DIR/install

ADD ./.rpm/dist/*_docker-el7.x86_64.rpm $TRUSTY_DIR/install/
RUN yum -y localinstall $(find $TRUSTY_DIR/install -name "*.rpm" | head -n 1)
RUN rm -rf $TRUSTY_DIR/install/

VOLUME ["/var/trusty/roots", "/var/trusty/certs", "/var/trusty/logs", "/var/trusty/audit"]

EXPOSE 8080 7891 7892 7893

# Define default command.
CMD ["/opt/trusty/bin/trusty"]