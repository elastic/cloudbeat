FROM debian
RUN apt-get update
RUN apt-get install -y ca-certificates
RUN apt-get install -y curl
RUN apt-get install -y iputils-ping
RUN apt-get install -y jq
COPY ./cloudbeat /cloudbeat
COPY ./cloudbeat.yml /cloudbeat.yml
ENTRYPOINT ["/cloudbeat"]
CMD ["-e", "-d", "'*'"]
