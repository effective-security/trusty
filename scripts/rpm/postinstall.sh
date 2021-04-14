#!/bin/bash

getent group trusty || groupadd -g 12000 -r trusty
getent passwd trusty || useradd -r -u 310000 -d /home/trusty -s /sbin/nologin -g trusty trusty

mkdir -p /var/trusty/data \
         /var/trusty/logs \
         /var/trusty/audit \
         /var/trusty/certs \
         /var/trusty/softhsm/tokens

chown -R trusty:trusty /var/trusty
chmod 0700 /var/trusty/data

systemctl daemon-reload
systemctl enable trusty
