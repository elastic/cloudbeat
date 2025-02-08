ARG ELASTIC_AGENT_IMAGE=docker.elastic.co/beats/elastic-agent:8.14.0-SNAPSHOT

FROM ${ELASTIC_AGENT_IMAGE} as elastic_agent_cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=755 cloudbeat /tmp/components/cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=666 bundle.tar.gz /tmp/components/bundle.tar.gz
COPY --chown=elastic-agent:elastic-agent --chmod=644 cloudbeat.yml /tmp/components/cloudbeat.yml
COPY --chown=elastic-agent:elastic-agent --chmod=644 cloudbeat.spec.yml /tmp/components/cloudbeat.spec.yml

RUN mv /tmp/components/* "$(dirname "$(realpath "$(which elastic-agent)")")"/components/

FROM elastic_agent_cloudbeat
