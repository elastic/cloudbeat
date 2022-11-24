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

## Cloudbeat Development

### Table of contents

- [Prerequisites](#prerequisites)
- [Deploying Cloudbeat as a process](#deploying-cloudbeat)
    - [Unmanaged Kubernetes](#clean-up)
    - [Amazon Elastic Kubernetes Service (EKS)](#amazon-elastic-kubernetes-service-(EKS))
- [Deploying Cloudbeat with Elastic-Agent](#running-cloudbeat-with-elastic-agent)

## Prerequisites

1. We use [Hermit](https://cashapp.github.io/hermit/usage/get-started/) to keep all our tooling in check. See our [README](bin/README.hermit.md) for more details.
   Install it
   with the following command:
    ```zsh
    curl -fsSL https://github.com/cashapp/hermit/releases/download/stable/install.sh | /bin/bash
    . ./bin/activate-hermit
    ```

   > **Note**
   This will download and install hermit into `~/bin`. You should add this to your `$PATH` if it isn't already.

   It is also recommended to add hermit's [shell integration](https://cashapp.github.io/hermit/usage/shell/)

2. Elastic stack running locally, preferably using [Elastic-Package](https://github.com/elastic/elastic-package) (you
   may need to [authenticate](https://docker-auth.elastic.co/github_auth))
   For example, spinning up 8.5.0 stack locally:

    ```zsh
    eval "$(elastic-package stack shellinit)" # load stack environment variables
    elastic-package stack up --version 8.5.0 -v -d
    ```

- _optional:_ Create local kind cluster to test against
  ```zsh
  just create-kind-cluster
  just elastic-stack-connect-kind # connect it to local elastic stack
  ```

# Deploying Cloudbeat

## Running Cloudbeat as a process

### Self-Managed Kubernetes
Build and deploying cloudbeat into your local kind cluster:

```zsh
just build-deploy-cloudbeat
```

### Amazon Elastic Kubernetes Service (EKS)

Export AWS creds as env vars, kustomize will use these to populate your cloudbeat deployment.

```zsh
export AWS_ACCESS_KEY="<YOUR_AWS_KEY>"
export AWS_SECRET_ACCESS_KEY="<YOUR_AWS_SECRET>"
```

Set your default cluster to your EKS cluster

```zsh
kubectl config use-context {your-eks-cluster}
```

Deploy cloudbeat on your EKS cluster

```zsh
just deploy-eks-cloudbeat
````

### Advanced

If you need to change the default values in the configuration(`ES_HOST`, `ES_PORT`, `ES_USERNAME`, `ES_PASSWORD`), you
can
also create the deployment file yourself.

Self-Managed Kubernetes
```zsh
just create-vanilla-deployment-file
```

EKS

```zsh
just create-eks-deployment-file
```

To validate check the logs:

```zsh
just logs-cloudbeat
```

Now go and check out the data on your Kibana!

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

> **Note**
> Use your favorite IDE to connect to the debugger on `localhost:40000` (for
> example [Goland](https://www.jetbrains.com/help/go/attach-to-running-go-processes-with-debugger.html#step-3-create-the-remote-run-debug-configuration-on-the-client-computer))

## Running Cloudbeat with Elastic Agent

Cloudbeat is only supported on managed Elastic-Agents. It means, that in order to run the setup, you will be required to
have a Kibana running.
Create an agent policy and install the CSP integration. Now, when adding a new agent, you will get the K8s deployment
instructions of elastic-agent.

> **Note** Are you a developer/contributor or just looking for more deployment types? check out
> our [dev docs](dev-docs/Development.md)
