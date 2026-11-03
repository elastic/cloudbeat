ARG ELASTIC_AGENT_IMAGE=docker.elastic.co/beats/elastic-agent:8.19.24-SNAPSHOT@sha256:ff4738548188dabebcc47948c9adc3c99c1657bcf9565d53b37ecc914a1de9dc

FROM ${ELASTIC_AGENT_IMAGE} as elastic_agent_cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=755 cloudbeat /tmp/components/cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=666 bundle.tar.gz /tmp/components/bundle.tar.gz
COPY --chown=elastic-agent:elastic-agent --chmod=644 cloudbeat.yml /tmp/components/cloudbeat.yml

RUN mv /tmp/components/* "$(dirname "$(realpath "$(which elastic-agent)")")"/components/

FROM elastic_agent_cloudbeat
