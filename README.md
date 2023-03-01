[![Coverage Status](https://coveralls.io/repos/github/elastic/cloudbeat/badge.svg?branch=main)](https://coveralls.io/github/elastic/cloudbeat?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/elastic/cloudbeat)](https://goreportcard.com/report/github.com/elastic/cloudbeat)

# Cloudbeat

Cloudbeat analyzes cloud assets for security compliance and sends findings to Elasticsearch as part of
the [Cloud Security Posture](https://www.elastic.co/blog/secure-your-cloud-with-elastic-security) plugin in Kibana.

## Getting Started

To get started with Cloud Security Posture on your cluster, see
our [documentation](https://www.elastic.co/guide/en/security/master/get-started-with-kspm.html#kspm-setup-unmanaged).

- [Setup KSPM for Amazon EKS clusters](https://www.elastic.co/guide/en/security/master/get-started-with-kspm.html#kspm-setup-eks-start)

- [Setup KSPM for unmanaged Kubernetes clusters](https://www.elastic.co/guide/en/security/master/get-started-with-kspm.html#kspm-setup-unmanaged)

---

## Deployment

In order to run Cloudbeat, you need to have Elastic Stack (Elasticsearch, Kibana, etc) running (locally/cloud). See **[ELK Deployment options](dev-docs/ELK-Deployment.md)**

After deploying your Elastic Stack, you can deploy Cloudbeat. See **[Cloudbeat Deployment options](dev-docs/Cloudbeat-Deployment.md)**

### Devs Prerequisites

We use [Hermit](https://cashapp.github.io/hermit/usage/get-started/) to keep all our tooling in check. See our [README](/bin/README.hermit.md) for more details.

Install it with the following commands:
```zsh
curl -fsSL https://github.com/cashapp/hermit/releases/download/stable/install.sh | /bin/bash
. ./bin/activate-hermit
```
> **Note**
> This will download and install hermit into `~/bin`. You should add this to your `$PATH` if it isn't already.
It is also recommended to add hermit's [shell integration](https://cashapp.github.io/hermit/usage/shell/)


> **Note** Are you a developer/contributor or just looking for more information check out
> our [dev docs](dev-docs/Development.md)
