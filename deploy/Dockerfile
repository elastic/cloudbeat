FROM debian@sha256:e2d08da6f42ef4b09b165d55528a12727aeed8240dc9edf888e3ec07e10ef9da
RUN set -x && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
      ca-certificates curl iputils-ping jq && \
      apt-get clean

COPY ./cloudbeat /cloudbeat
COPY ./cloudbeat.yml /cloudbeat.yml
COPY ./bundle.tar.gz /bundle.tar.gz

ENTRYPOINT ["/cloudbeat"]
CMD ["-e", "-d", "'*'"]
