# Cloudbeat 
![Coverage](https://img.shields.io/badge/Coverage-46.4%25-yellow)
[![Go Report Card](https://goreportcard.com/badge/github.com/elastic/cloudbeat)](https://goreportcard.com/report/github.com/elastic/cloudbeat)
[![Build Status](https://internal-ci.elastic.co/buildStatus/icon?job=cloudbeat%2Fcloudbeat-mbp%2Fmain)](https://internal-ci.elastic.co/job/cloudbeat/job/cloudbeat-mbp/job/main/)

## Table of contents
- [Prerequisites](#prerequisites)
- [Running Cloudbeat](#running-cloudbeat)
  - [Clean up](#clean-up)
  - [Remote Debugging](#remote-debugging)


## Prerequisites
1. [Just command runner](https://github.com/casey/just)
2. Elasticsearch with the default username & password (`elastic` & `changeme`) running on the default port (`http://localhost:9200`)
3. Kibana with running on the default port (`http://localhost:5601`)
4. Set up the local env:

```zsh
just setup-env
```

## Running Cloudbeat

Build & deploy cloudbeat:

```zsh
just build-deploy-cloudbeat
```

To validate check the logs:

```zsh
just logs-cloudbeat
```

Now go and check out the data on your Kibana! Make sure to add a kibana dataview `logs-cis_kubernetes_benchmark.findings-*`

### Clean up

To stop this example and clean up the pod, run:
```zsh
just delete-cloudbeat
```
### Remote Debugging

Build & Deploy remote debug docker:

```zsh
just build-deploy-cloudbeat-debug
```

After running the pod, expose the relevant ports:
```zsh
just expose-ports
```

The app will wait for the debugger to connect before starting

```zsh
just logs-cloudbeat
```

Use your favorite IDE to connect to the debugger on `localhost:40000` (for example [Goland](https://www.jetbrains.com/help/go/attach-to-running-go-processes-with-debugger.html#step-3-create-the-remote-run-debug-configuration-on-the-client-computer))

Note: Check the jusfile for all available commands for build or deploy `$ just --summary`
</br>

