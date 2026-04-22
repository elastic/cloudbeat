ARG ELASTIC_AGENT_IMAGE=docker.elastic.co/beats/elastic-agent:8.19.15-SNAPSHOT@sha256:82ea0467d47439b392d1d6ae2737245feb6789980b61e97cf7a0473da983debe

FROM ${ELASTIC_AGENT_IMAGE} AS elastic_agent_cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=755 cloudbeat /tmp/components/cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=666 bundle.tar.gz /tmp/components/bundle.tar.gz
COPY --chown=elastic-agent:elastic-agent --chmod=644 cloudbeat.yml /tmp/components/cloudbeat.yml

RUN mv /tmp/components/* "$(dirname "$(readlink -f elastic-agent)")"/components/

FROM elastic_agent_cloudbeat
