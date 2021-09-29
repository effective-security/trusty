FROM ekspand/trusty-docker-base:latest
LABEL org.opencontainers.image.authors="Martini Security <denis@martinisecurity.com>" \
      org.opencontainers.image.url="https://github.com/martinisecurity/trusty" \
      org.opencontainers.image.source="https://github.com/martinisecurity/trusty" \
      org.opencontainers.image.documentation="https://github.com/martinisecurity/trusty" \
      org.opencontainers.image.vendor="Martini Security" \
      org.opencontainers.image.description="Trusty CA"

ENV PATH=$PATH:/opt/trusty/bin

RUN mkdir -p /home/nonroot /opt/trusty/bin /opt/trusty/sql /var/trusty/certs /var/trusty/audit /opt/trusty/etc/prod /opt/trusty/etc/dev /trusty_certs
COPY ./bin/trusty* /opt/trusty/bin/
COPY ./scripts/build/* /opt/trusty/bin/
COPY ./sql/ /opt/trusty/sql/

VOLUME ["/var/trusty/certs", \
"/var/trusty/logs", \
"/var/trusty/audit", \
"/opt/trusty/sql", \
"/opt/trusty/etc/prod", \
"/opt/trusty/etc/dev"]

EXPOSE 7880 7892

RUN groupadd -g 1000 -o nonroot
RUN useradd -r -u 1000 -g nonroot nonroot
RUN chown -R nonroot:nonroot /home/nonroot 
RUN chown -R nonroot:nonroot /var/trusty /trusty_certs

USER nonroot:nonroot

# Define default command.
CMD ["/opt/trusty/bin/trusty"]