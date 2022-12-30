[![Coverage Status](https://coveralls.io/repos/github/elastic/cloudbeat/badge.svg?branch=main)](https://coveralls.io/github/elastic/cloudbeat?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/elastic/cloudbeat)](https://goreportcard.com/report/github.com/elastic/cloudbeat)

# Cloudbeat

Cloudbeat analyzes cloud assets for security compliance and sends findings to Elasticsearch as part of
the [Cloud Security Posture](https://www.elastic.co/blog/secure-your-cloud-with-elastic-security) plugin in Kibana.

## Getting Started

To get started with Cloud Security Posture on your cluster, see
our [documentation](https://www.elastic.co/guide/en/security/master/get-started-with-kspm.html#kspm-setup-unmanaged).

- [Setup KSPM for Amazon EKS clusters](https://www.elastic.co/guide/en/security/master/get-started-with-kspm.html#kspm-setup-unmanaged)

- [Setup KSPM for unmanaged Kubernetes clusters](https://www.elastic.co/guide/en/security/master/get-started-with-kspm.html#kspm-setup-eks-start)

___

## Table of contents

- [Local Deployment](#local-deployment)
    - [Unmanaged Kubernetes](#deploying-cloudbeat-as-a-process)
      - [Self-Managed Kubernetes (Kind)](#self-managed-kubernetes)
      - [Amazon Elastic Kubernetes Service (EKS)](#amazon-elastic-kubernetes-service-(EKS))
- [Deploying Cloudbeat with Elastic-Agent](#running-cloudbeat-with-elastic-agent)

## Local Deployment

Deploying Cloudbeat locally either as a process, or through elastic-agent can be done with [elastic-package](https://github.com/elastic/elastic-package) - a tool that spins up en entire elastic stack locally.
depending on the deployment platform (Self-Managed kubernetes / EKS) you may need to set up different environment.

### Prerequisites

1. We use [Hermit](https://cashapp.github.io/hermit/usage/get-started/) to keep all our tooling in check. See our [README](/bin/README.hermit.md) for more details.
   Install it with the following commands:
    ```zsh
    curl -fsSL https://github.com/cashapp/hermit/releases/download/stable/install.sh | /bin/bash
    . ./bin/activate-hermit
    ```

   > **Note**
   This will download and install hermit into `~/bin`. You should add this to your `$PATH` if it isn't already.

   It is also recommended to add hermit's [shell integration](https://cashapp.github.io/hermit/usage/shell/)

2. Elastic stack running locally, preferably using [elastic-package](https://github.com/elastic/elastic-package) (you
   may need to [authenticate](https://docker-auth.elastic.co/github_auth))
   For example, spinning up 8.6.0 stack locally:

    ```zsh
    eval "$(elastic-package stack shellinit --shell $(basename $SHELL))" # load stack environment variables
    elastic-package stack up --version 8.6.0 -v -d
    ```

## Deploying Cloudbeat as a process

Cloudbeat can be deployed as a process, and will not be managed by Elastic Agent. (the fastest way to get started, getting findings)

### Self-Managed Kubernetes

We use [Kind](https://kind.sigs.k8s.io/) to spin up a local kubernetes cluster, and deploy Cloudbeat as a process.
Build and deploying cloudbeat into your local kind cluster:

if you don't already have a Kind cluster, you can create one with:

```zsh
just create-kind-cluster
just elastic-stack-connect-kind # connect it to local elastic stack
```

Build and deploy cloudbeat on your local kind cluster:

```zsh
just build-deploy-cloudbeat
```

> **Note** By default, cloudbeat binary will be built based on `GOOS` and `GOARCH` environment variables.
> If you want to build cloudbeat for a different platform you can set them as following:
> ```zsh
> # just build-deploy-cloudbeat <Target OS> <Target Arch>
> just build-deploy-cloudbeat linux amd64
> ```

### Amazon Elastic Kubernetes Service (EKS)

Another deployment option is to deploy cloudbeat as a process on EKS. This is useful for testing and development purposes.

Export AWS creds as env vars, kustomize will use these to populate your cloudbeat deployment.

```zsh
export AWS_ACCESS_KEY="<YOUR_AWS_KEY>"
export AWS_SECRET_ACCESS_KEY="<YOUR_AWS_SECRET>"
```

Set your default cluster to your EKS cluster

```zsh
kubectl config use-context <your-eks-cluster>
```

Deploy cloudbeat on your EKS cluster

```zsh
just deploy-eks-cloudbeat
````

## Running Cloudbeat with Elastic Agent


1. Spin up Elastic stack (using [cloud](https://staging.found.no/home)/[staging](https://staging.found.no/home) is recommended, but using elastic-package is also supported, see [Local Deployment](#local-deployment))
2. Create an agent policy and install the CSP integration (KSPM).
3. Now, when adding a new agent, you will get the K8s deployment instructions of elastic-agent.

> **Note** Are you a developer/contributor or just looking for more information check out
> our [dev docs](dev-docs/Development.md)
