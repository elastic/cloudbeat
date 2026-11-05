ARG ELASTIC_AGENT_IMAGE=docker.elastic.co/beats/elastic-agent:8.19.25-SNAPSHOT@sha256:e5256c1fbe4d44fe9f672dcf64fa18124d1f9a9aecdc564bab07b8d09ece0b4f

FROM ${ELASTIC_AGENT_IMAGE} as elastic_agent_cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=755 cloudbeat /tmp/components/cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=666 bundle.tar.gz /tmp/components/bundle.tar.gz
COPY --chown=elastic-agent:elastic-agent --chmod=644 cloudbeat.yml /tmp/components/cloudbeat.yml

RUN mv /tmp/components/* "$(dirname "$(realpath "$(which elastic-agent)")")"/components/

FROM elastic_agent_cloudbeat
