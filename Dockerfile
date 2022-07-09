FROM ekspand/trusty-docker-base:latest
LABEL org.opencontainers.image.authors="Effective Security <denis@effectivesecurity.pt>" \
      org.opencontainers.image.url="https://github.com/effective-security/trusty" \
      org.opencontainers.image.source="https://github.com/effective-security/trusty" \
      org.opencontainers.image.documentation="https://github.com/effective-security/trusty" \
      org.opencontainers.image.vendor="Effective Security" \
      org.opencontainers.image.description="Trusty CA"

ENV PATH=$PATH:/opt/trusty/bin

RUN mkdir -p /home/nonroot \
      /opt/trusty/bin /opt/trusty/sql /opt/trusty/etc/prod /opt/trusty/etc/dev \
      /var/trusty/certs /var/trusty/logs \
      /trusty_certs /trusty_logs
COPY ./bin/trusty* ./bin/hsm-tool ./bin/xpki-tool /opt/trusty/bin/
COPY ./scripts/build/* ./change_log.txt /opt/trusty/bin/
COPY ./sql/ /opt/trusty/sql/

VOLUME ["/var/trusty/certs", \
      "/var/trusty/logs", \
      "/opt/trusty/sql", \
      "/opt/trusty/etc/prod", \
      "/opt/trusty/etc/dev"]

EXPOSE 7880 7892 9090

RUN groupadd -g 1000 -o nonroot
RUN useradd -r -u 1000 -g nonroot nonroot
RUN chown -R nonroot:nonroot /home/nonroot 
RUN chown -R nonroot:nonroot /var/trusty /trusty_certs /trusty_logs

USER nonroot:nonroot

# Define default command.
CMD ["/opt/trusty/bin/trusty"]