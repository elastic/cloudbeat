FROM debian@sha256:34cd9e9fd437c0a095ec39cb2e73422c9f30821b0d0848ed74fd0d43bae4d958
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
