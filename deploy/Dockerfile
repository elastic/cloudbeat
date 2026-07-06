FROM debian@sha256:35b8ff74ead4880f22090b617372daff0ccae742eb5674455d542bef71ef1999
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
