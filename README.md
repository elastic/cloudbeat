# Cloudbeat
[![Coverage Status](https://coveralls.io/repos/github/elastic/cloudbeat/badge.svg?branch=main)](https://coveralls.io/github/elastic/cloudbeat?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/elastic/cloudbeat)](https://goreportcard.com/report/github.com/elastic/cloudbeat)

Cloudbeat evaluates cloud assets for security compliance and ships findings to Elasticsearch

### Table of contents
- [Prerequisites](#prerequisites)
- [Deploying Cloudbeat as a process](#deploying-cloudbeat)
  - [Unmanaged Kubernetes](#clean-up)
  - [EKS](#remote-debugging)
- [Deploying Cloudbeat with Elastic-Agent](#running-cloudbeat-with-elastic-agent)


# Prerequisites
[Hermit](https://cashapp.github.io/hermit/usage/get-started/)

- Install & activate hermit

  ```zsh
  curl -fsSL https://github.com/cashapp/hermit/releases/download/stable/install.sh | /bin/bash
  . ./bin/activate-hermit
  ```

  >  **Note**
  This will download and install hermit into `~/bin`. You should add this to your `$PATH` if it isn't already.

- _optional:_ Create local kind cluster
  ```zsh
  just create-kind-cluster
  just elastic-stack-connect-kind # connect it to local elastic stack
  ```

- Elastic stack running locally, preferably using [Elastic-Package](https://github.com/elastic/elastic-package) (you may need to [authenticate](https://docker-auth.elastic.co/github_auth))

  For example, spinning up 8.5.0 stack locally:

  ```zsh
  eval "$(elastic-package stack shellinit)" # load stack environment variables
  elastic-package stack up --version 8.5.0 -v -d
  ```


# Deploying Cloudbeat
## Running Cloudbeat as a process
### Unmanaged Kubernetes (Vanilla)
Build & deploy cloudbeat:

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

If you need to change the default values in the configuration(`ES_HOST`, `ES_PORT`, `ES_USERNAME`, `ES_PASSWORD`), you can
also create the deployment file yourself.

Vanilla
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

>**Note**
Use your favorite IDE to connect to the debugger on `localhost:40000` (for example [Goland](https://www.jetbrains.com/help/go/attach-to-running-go-processes-with-debugger.html#step-3-create-the-remote-run-debug-configuration-on-the-client-computer))


## Running Cloudbeat with Elastic Agent
Cloudbeat is only supported on managed elastic-agents. It means, that in order to run the setup, you will be required to have a Kibana running.
Create an agent policy and install the CSP integration. Now, when adding a new agent, you will get the K8s deployment instructions of elastic-agent.

### Update settings
Update cloudbeat settings on a running elastic-agent can be done by running the [script](/scripts/remote_edit_config.sh).
The script still requires a second step of trigerring the agent to re-run cloudbeat.
This can be done on Fleet UI by changing the agent log level.
Another option is through CLI on the agent by running
```
kill -9 `pidof cloudbeat`
```

### Local configuration changes
To update your local configuration of cloudbeat and control it, use
```sh
mage config
```

In order to control the policy type you can pass the following environment variable
```sh
POLICY_TYPE=cloudbeat/cis_eks mage config
```

The default `POLICY_TYPE` is set to `cloudbeat/cis_k8s` on [`_meta/config/cloudbeat.common.yml.tmpl`](_meta/config/cloudbeat.common.yml.tmpl)


