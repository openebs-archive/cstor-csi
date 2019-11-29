#
# This Dockerfile builds a recent volume-mgmt using the latest binary from
# volume-mgmt  releases.
#

FROM ubuntu:16.04
RUN apt-get update; exit 0
RUN apt-get -y install rsyslog xfsprogs
#RUN apt-get clean && rm -rf /var/lib/apt/lists/*

COPY cstor-csi-driver /usr/local/bin/
COPY entrypoint.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/entrypoint.sh

ARG BUILD_DATE
LABEL org.label-schema.name="csi-driver"
LABEL org.label-schema.description="OpenEBS"
LABEL org.label-schema.url="http://www.openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/cstor-csi"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT ["/usr/local/bin/cstor-csi-driver"]
EXPOSE 7676 7777
