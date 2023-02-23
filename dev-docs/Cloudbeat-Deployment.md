# Cloudbeat Deployment

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


> **Note** By default, cloudbeat binary will be built based on `GOARCH` environment variable.
> If you want to build cloudbeat for a different platform you can set it as following:
>
> ```zsh
> # just build-deploy-cloudbeat <Target Arch>
> just build-deploy-cloudbeat amd64
> ```
>
> Or without certificate
>
> ```zsh
> # just build-deploy-cloudbeat-nocert <Target Arch>
> just build-deploy-cloudbeat-nocert amd64
> ```


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


## Deploying Cloudbeat with standalone Elastic Agent

1. Spin up Elastic stack (See [ELK stack setup](ELK-Deployment.md))
2. Collect the relevant information from the Fleet UI:
   - Fleet URL
   - Enrollment token
3. Use docker to run the standalone agent, for exmaple:
   ```zsh
   docker run -d --platform=linux/x86_64 -e "FLEET_URL=<FLEET_URL>" -e "FLEET_ENROLLMENT_TOKEN=<FLEET_ENROLLMENT_TOKEN>" -e "FLEET_ENROLL=1" docker.elastic.co/beats/elastic-agent:8.7.0-SNAPSHOT
   ```
