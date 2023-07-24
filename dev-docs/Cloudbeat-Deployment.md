# Cloudbeat Deployment

## Table of Contents
- [Deploying Cloudbeat as a process](#deploying-cloudbeat-as-a-process)
   - [Self-Managed Kubernetes](#self-managed-kubernetes)
   - [Amazon Elastic Kubernetes Service (EKS)](#amazon-elastic-kubernetes-service-eks)
- [Deploying Cloudbeat with managed Elastic Agent](#deploying-cloudbeat-with-managed-elastic-agent)
- [Deploying Cloudbeat with standalone Elastic Agent](#deploying-cloudbeat-with-standalone-elastic-agent)

## Deploying Cloudbeat as a process

Cloudbeat can be deployed as a process, and will not be managed by Elastic Agent. (the fastest way to get started, getting findings)

### Self-Managed Kubernetes

We use [Kind](https://kind.sigs.k8s.io/) to spin up a local kubernetes cluster, and deploy Cloudbeat as a process.
Build and deploying cloudbeat into your local kind cluster:

1. if you don't already have a Kind cluster, you can create one with:

   ```zsh
   just create-kind-cluster
   just elastic-stack-connect-kind # connect it to local elastic stack
   ```

2. Build and deploy cloudbeat on your local kind cluster:

   ```zsh
   just build-deploy-cloudbeat
   ```

3. Or without certificate

   ```zsh
   just build-deploy-cloudbeat-nocert
   ```

#### Cross-platform Builds

By default, cloudbeat binary will be built based on `GOARCH` environment variable.
If you want to build cloudbeat for a different platform you can set it as following:

```zsh
# just build-deploy-cloudbeat <Target Arch>
just build-deploy-cloudbeat amd64
```

Or without certificate

```zsh
# just build-deploy-cloudbeat-nocert <Target Arch>
just build-deploy-cloudbeat-nocert amd64
```

#### Stopping / Restarting the local Elastic Stack

If you are using `elastic-package` to run the Elastic Stack locally and need to take it down
with `elastic-package stack down`, you might run into errors:
```shell
failed to remove network elastic-package-stack_default: Error response from daemon: error while removing network: network elastic-package-stack_default id <id> has active endpoints
Error: tearing down the stack failed: stopping docker containers failed: running command failed: running Docker Compose down command failed: exit status 1
```

You can fix this by disconnecting the kind cluster from the stack with:

```zsh
just elastic-stack-disconnect-kind
```

### Amazon Elastic Kubernetes Service (EKS)

Another deployment option is to deploy cloudbeat as a process on a managed Kubernetes cluster (EKS in our case).
This is useful for testing and development purposes.

1. Export AWS creds as env vars, Kustomize will use these to populate your cloudbeat deployment.

   ```zsh
   export AWS_ACCESS_KEY="<YOUR_AWS_KEY>"
   export AWS_SECRET_ACCESS_KEY="<YOUR_AWS_SECRET>"
   ```

2. Set your default cluster to your EKS cluster

   ```zsh
   kubectl config use-context <your-eks-cluster>
   ```

3. Deploy cloudbeat on your EKS cluster

   ```zsh
   just deploy-eks-cloudbeat
   ```


## Deploying Cloudbeat with managed Elastic Agent

1. Spin up Elastic stack (See [ELK stack setup](ELK-Deployment.md))
2. Create an agent policy and install the CSPM/KSPM integration.
3. Now, when adding a new agent, you will get the K8s deployment instructions of elastic-agent.
   - For KSPM it's recommended to use the `DaemonSet` deployment.
   - For CSPM it's recommended to use the run the agent as a linux binary (darwin is not supported yet).


## Deploying Fleet enrolled Elastic Agent in a container

1. Spin up Elastic stack (See [ELK stack setup](ELK-Deployment.md))
2. Collect the relevant information from the Fleet UI:
   - Fleet URL
   - Enrollment token
3. It's recommended to use docker to run the standalone agent, for example:
   ```zsh
   docker run -d --platform=linux/x86_64 \
   -e "FLEET_URL=<fleet-server-host-url>" \
   -e "FLEET_ENROLLMENT_TOKEN=<enrollment-token>" \
   -e "FLEET_ENROLL=1" \
   docker.elastic.co/beats/elastic-agent:8.7.0-SNAPSHOT
   ```

For more information see [Run Elastic Agent in a container](https://www.elastic.co/guide/en/fleet/current/elastic-agent-container.html#elastic-agent-container).
