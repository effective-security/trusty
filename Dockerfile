FROM centos:7

ENV TRUSTY_DIR=/opt/trusty
ENV PATH=$PATH:/opt/trusty/bin
RUN mkdir -p $TRUSTY_DIR/install

RUN yum install -y https://yum.postgresql.org/11/redhat/rhel-7-x86_64/pgdg-redhat-repo-latest.noarch.rpm

RUN yum update -y && \
    yum install -y ca-certificates \
        which \
        opensc \
        softhsm \
        postgresql11 \
        libtool-ltdl-devel

ADD ./.rpm/dist/*_docker-el7.x86_64.rpm $TRUSTY_DIR/install/
RUN yum -y localinstall $(find $TRUSTY_DIR/install -name "*.rpm" | head -n 1)
RUN rm -rf $TRUSTY_DIR/install/

VOLUME ["/var/trusty/roots", "/var/trusty/certs", "/var/trusty/logs", "/var/trusty/audit"]

EXPOSE 8080 7891

# Define default command.
CMD ["/opt/trusty/bin/trusty"]