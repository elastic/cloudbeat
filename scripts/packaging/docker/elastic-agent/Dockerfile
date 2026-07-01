ARG ELASTIC_AGENT_IMAGE=docker.elastic.co/beats/elastic-agent:8.19.18-SNAPSHOT@sha256:0d6735edf18e33e0658079eb54d61fde6170cd90290f1c5f9160bca2f14be5a0

FROM ${ELASTIC_AGENT_IMAGE} AS elastic_agent_cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=755 cloudbeat /tmp/components/cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=666 bundle.tar.gz /tmp/components/bundle.tar.gz
COPY --chown=elastic-agent:elastic-agent --chmod=644 cloudbeat.yml /tmp/components/cloudbeat.yml

RUN mv /tmp/components/* "$(dirname "$(readlink -f elastic-agent)")"/components/

FROM elastic_agent_cloudbeat
