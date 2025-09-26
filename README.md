[![Coverage Status](https://coveralls.io/repos/github/elastic/cloudbeat/badge.svg?branch=main)](https://coveralls.io/github/elastic/cloudbeat?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/elastic/cloudbeat)](https://goreportcard.com/report/github.com/elastic/cloudbeat)
[![Build status](https://badge.buildkite.com/82f39bb3a95eeb7f46e28891fb48a623cf184fbfca2eff545a.svg)](https://buildkite.com/elastic/cloudbeat)

# Cloudbeat

Cloudbeat is a tool that analyzes cloud assets for security compliance and sends findings to Elasticsearch.
It is designed to be used as part of the [Cloud Security](https://www.elastic.co/blog/secure-your-cloud-with-elastic-security) plugin in Kibana.


### CSP Security Policies

Cloudbeat uses [security policies](./security-policies) to evaluate cloud resources.

## Getting Started

To get started with Cloud Security on your cluster, please refer to our documentation:

- [Get started with Kubernetes Security Posture Management (KSPM)](https://www.elastic.co/guide/en/security/current/get-started-with-kspm.html)

- [Get started with Cloud Security Posture Management (CSPM)](https://www.elastic.co/guide/en/security/current/cspm.html)

- [Get started with Cloud Native Vulnerability Management (CNVM)](https://www.elastic.co/guide/en/security/current/vuln-management-get-started.html)

---

## Deployment

To run Cloudbeat, you need to have Elastic Stack (Elasticsearch, Kibana, etc) running (locally/cloud). See our [Elastic Stack Deployment options](dev-docs/ELK-Deployment.md) documentation.

Once your Elastic Stack is deployed, you can proceed with the deployment of Cloudbeat. For deployment instructions, see [Cloudbeat Deployment options](dev-docs/Cloudbeat-Deployment.md).

## Development

### Prerequisites

We use [Hermit](https://cashapp.github.io/hermit/usage/get-started/) to manage our development tooling. Please refer to our [README](/bin/README.hermit.md) for detailed instructions on setting it up.

___

> **Note** If you are a developer or contributor, or if you are looking for additional information, please visit our [development documentation](dev-docs/Development.md)
