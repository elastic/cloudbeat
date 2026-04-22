ARG ELASTIC_AGENT_IMAGE=docker.elastic.co/beats/elastic-agent:9.0.0-SNAPSHOT@sha256:2edbab5551a817c9a7532f8294477fb834bc7006e4e90eb0b70a52271d118d9c

FROM ${ELASTIC_AGENT_IMAGE} AS elastic_agent_cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=755 cloudbeat /tmp/components/cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=666 bundle.tar.gz /tmp/components/bundle.tar.gz
COPY --chown=elastic-agent:elastic-agent --chmod=644 cloudbeat.yml /tmp/components/cloudbeat.yml

RUN mv /tmp/components/* "$(dirname "$(readlink -f elastic-agent)")"/components/

FROM elastic_agent_cloudbeat
