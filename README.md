[![Coverage Status](https://coveralls.io/repos/github/elastic/cloudbeat/badge.svg?branch=main)](https://coveralls.io/github/elastic/cloudbeat?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/elastic/cloudbeat)](https://goreportcard.com/report/github.com/elastic/cloudbeat)
[![Build status](https://badge.buildkite.com/82f39bb3a95eeb7f46e28891fb48a623cf184fbfca2eff545a.svg)](https://buildkite.com/elastic/cloudbeat)

# Cloudbeat

Cloudbeat is a tool that analyzes cloud assets for security compliance and sends findings to Elasticsearch. 
It is designed to be used as part of the [Cloud Security Posture](https://www.elastic.co/blog/secure-your-cloud-with-elastic-security) plugin in Kibana.


### CSP Security Policies

Cloudbeat uses security policies from the [CSP Security Policies](https://github.com/elastic/csp-security-policies) repository to evaluate cloud resources.

## Getting Started

To get started with Cloud Security Posture on your cluster, please refer to our documentation:

- [Get started with KSPM (Kubernetes Security Posture Management)](https://www.elastic.co/guide/en/security/master/get-started-with-kspm.html#get-started-with-kspm)

- [Get started with CSPM (Cloud Security Posture Management)](https://www.elastic.co/guide/en/security/master/cspm-get-started.html#cspm-get-started)

- [Get started with CNVM (Cloud Native Vulnerability Management)](https://www.elastic.co/guide/en/security/master/vuln-management-overview.html)

---

## Deployment

To run Cloudbeat, you need to have Elastic Stack (Elasticsearch, Kibana, etc) running (locally/cloud). See **[Elastic Stack Deployment options](dev-docs/ELK-Deployment.md)**

After deploying your Elastic Stack, you can deploy Cloudbeat. See **[Cloudbeat Deployment options](dev-docs/Cloudbeat-Deployment.md)**

## Development

### Prerequisites

We use [Hermit](https://cashapp.github.io/hermit/usage/get-started/) to keep all our tooling in check. See our [README](/bin/README.hermit.md) for more details.

Install it with the following commands:
```zsh
curl -fsSL https://github.com/cashapp/hermit/releases/download/stable/install.sh | /bin/bash
. ./bin/activate-hermit
```
> **Note**
> This will download and install hermit into `~/bin`. You should add this to your `$PATH` if it isn't already.
It is also recommended to add hermit's [shell integration](https://cashapp.github.io/hermit/usage/shell/)


If you are a developer or contributor, or if you are looking for additional information, please visit our [development documentation](dev-docs/Development.md)
