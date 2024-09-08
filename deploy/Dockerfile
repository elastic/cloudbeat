FROM debian
RUN set -x && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
      ca-certificates curl iputils-ping jq && \
      apt-get clean

COPY ./cloudbeat /cloudbeat
COPY ./cloudbeat.yml /cloudbeat.yml
COPY ./cloudbeat.spec.yml /cloudbeat.spec.yml
COPY ./bundle.tar.gz /bundle.tar.gz

ENTRYPOINT ["/cloudbeat"]
CMD ["-e", "-d", "'*'"]
