ARG ELASTIC_AGENT_IMAGE=docker.elastic.co/beats/elastic-agent:8.19.25-SNAPSHOT@sha256:e6ec65aecb16c41c9b5ae0a4eec0bb9eb45057934ab64157303838f55874e47f

FROM ${ELASTIC_AGENT_IMAGE} as elastic_agent_cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=755 cloudbeat /tmp/components/cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=666 bundle.tar.gz /tmp/components/bundle.tar.gz
COPY --chown=elastic-agent:elastic-agent --chmod=644 cloudbeat.yml /tmp/components/cloudbeat.yml

RUN mv /tmp/components/* "$(dirname "$(realpath "$(which elastic-agent)")")"/components/

FROM elastic_agent_cloudbeat
