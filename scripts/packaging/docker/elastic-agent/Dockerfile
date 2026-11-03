ARG ELASTIC_AGENT_IMAGE=docker.elastic.co/elastic-agent/elastic-agent:9.5.6@sha256:0852ec96b475d64a70d28644c4d898bd2c913d9962b3323998aa77f27e5393e5

FROM ${ELASTIC_AGENT_IMAGE} AS elastic_agent_cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=755 cloudbeat /tmp/components/cloudbeat
COPY --chown=elastic-agent:elastic-agent --chmod=666 bundle.tar.gz /tmp/components/bundle.tar.gz
COPY --chown=elastic-agent:elastic-agent --chmod=644 cloudbeat.yml /tmp/components/cloudbeat.yml

RUN mv /tmp/components/* "$(dirname "$(readlink -f elastic-agent)")"/components/

FROM elastic_agent_cloudbeat
