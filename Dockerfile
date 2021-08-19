FROM ekspand/trusty-docker-base:latest
LABEL org.opencontainers.image.source https://github.com/ekspand/trusty

LABEL org.opencontainers.image.authors="Martini Security <denis@martinisecurity.com>" \
      org.opencontainers.image.url="https://github.com/martinisecurity/trusty" \
      org.opencontainers.image.source="https://github.com/martinisecurity/trusty" \
      org.opencontainers.image.documentation="https://github.com/martinisecurity/trusty" \
      org.opencontainers.image.vendor="Martini Security" \
      org.opencontainers.image.description="Trusty CA"

ENV TRUSTY_DIR=/opt/trusty
ENV PATH=$PATH:$TRUSTY_DIR/bin

RUN mkdir -p $TRUSTY_DIR/bin $TRUSTY_DIR/sql
COPY ./bin/trusty* $TRUSTY_DIR/bin/
COPY ./scripts/build/* $TRUSTY_DIR/bin/
COPY ./sql/ $TRUSTY_DIR/sql/

VOLUME ["/var/trusty/certs", "/var/trusty/logs", "/var/trusty/audit"]

EXPOSE 7880 7891 7892

RUN groupadd -g 1000 -o nonroot
RUN useradd -r -u 1000 -g nonroot nonroot
RUN mkdir -p /home/nonroot/.config/softhsm2
RUN chown -R nonroot:nonroot /home/nonroot

USER nonroot:nonroot

# Define default command.
CMD ["$TRUSTY_DIR/bin/trusty"]