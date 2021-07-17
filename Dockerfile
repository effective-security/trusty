FROM ekspand/trusty-docker-base:latest
LABEL org.opencontainers.image.source https://github.com/ekspand/trusty

LABEL org.opencontainers.image.authors="Ekspand <denis@ekspand.com>" \
      org.opencontainers.image.url="https://github.com/ekspand/trusty" \
      org.opencontainers.image.source="https://github.com/ekspand/trusty" \
      org.opencontainers.image.documentation="https://github.com/ekspand/trusty" \
      org.opencontainers.image.vendor="Ekspand" \
      org.opencontainers.image.description="Trusty CA"

ENV TRUSTY_DIR=/opt/trusty
ENV PATH=$PATH:$TRUSTY_DIR/bin

RUN mkdir -p $TRUSTY_DIR/bin $TRUSTY_DIR/sql
COPY ./bin/trusty* $TRUSTY_DIR/bin/
COPY ./scripts/build/* $TRUSTY_DIR/bin/
COPY ./sql/ $TRUSTY_DIR/sql/

VOLUME ["/var/trusty/certs", "/var/trusty/logs", "/var/trusty/audit"]

EXPOSE 7880 7891 7892 7893

# Define default command.
CMD ["$TRUSTY_DIR/bin/trusty"]