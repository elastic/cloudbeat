ARG ELASTIC_AGENT_IMAGE=docker.elastic.co/beats/elastic-agent:8.19.25-SNAPSHOT@sha256:07be411dc617988280b1c30c39157cccb71e73974c052ca5c796021ec096e7ba

FROM ${ELASTIC_AGENT_IMAGE} as elastic_agent_cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=755 cloudbeat /tmp/components/cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=666 bundle.tar.gz /tmp/components/bundle.tar.gz
COPY --chown=elastic-agent:elastic-agent --chmod=644 cloudbeat.yml /tmp/components/cloudbeat.yml

RUN mv /tmp/components/* "$(dirname "$(realpath "$(which elastic-agent)")")"/components/

FROM elastic_agent_cloudbeat
