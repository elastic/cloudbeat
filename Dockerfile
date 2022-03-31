ARG GO_VERSION=1.17.2
FROM golang:${GO_VERSION} as builder
RUN apt-get update

# Setup work environment
ENV CLOUDBEAT_PATH /go/src/github.com/elastic/cloudbeat

RUN mkdir -p $CLOUDBEAT_PATH
WORKDIR $CLOUDBEAT_PATH

COPY . $CLOUDBEAT_PATH

RUN make

FROM golang:${GO_VERSION}
RUN set -x && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
      ca-certificates curl iputils-ping jq && \
      apt-get clean

ENV CLOUDBEAT_PATH /go/src/github.com/elastic/cloudbeat

COPY --from=builder $CLOUDBEAT_PATH/cloudbeat /cloudbeat
COPY --from=builder $CLOUDBEAT_PATH/cloudbeat.yml /cloudbeat.yml

CMD ./apm-server -e -d "*"
