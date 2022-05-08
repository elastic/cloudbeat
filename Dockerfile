FROM debian
RUN set -x && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
      ca-certificates curl iputils-ping jq && \
      apt-get clean

COPY ./cloudbeat /cloudbeat
COPY cloudbeat.yaml /cloudbeat.yaml

ENTRYPOINT ["/cloudbeat"]
CMD ["-e", "-d", "'*'"]
