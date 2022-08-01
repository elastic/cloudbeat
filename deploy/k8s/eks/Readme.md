# Setup Eks clusters with eksctl

[eksctl](https://eksctl.io/) is a simple CLI tool for creating and managing clusters on EKS - Amazon's managed Kubernetes service for EC2. It is written in Go, uses CloudFormation, was created by Weaveworks and it welcomes contributions from the community. Create a basic cluster in minutes with just one command.


## Installation

### Prerequisites
**Follow guidelines below if aws cli or kubectl are not yet installed on your local machine**.
See [references](#useful-references) section below for links to installation docs.

1. **AWS CLI**

   You will need to have AWS CLI credentials configured. You can use [`~/.aws/credentials` file][awsconfig]
   or [environment variables][awsenv]. For more information read [AWS documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html).

   [awsenv]: https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html
   [awsconfig]: https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html

   You will also need [AWS IAM Authenticator for Kubernetes](https://github.com/kubernetes-sigs/aws-iam-authenticator) command (either `aws-iam-authenticator` or `aws eks get-token` (available in version 1.16.156 or greater of AWS CLI) in your `PATH`.

2. **Kubectl**
    Once you have created a cluster, you will find that cluster credentials were added in `~/.kube/config`. If you have `kubectl` v1.10.x as well as `aws-iam-authenticator` commands in your PATH, you should be
    able to use `kubectl`. You will need to make sure to use the same AWS API credentials for this also. Check [EKS docs][ekskubectl] for instructions. If you installed `eksctl` via Homebrew, you should have all of these dependencies installed already.
    [ekskubectl]: https://docs.aws.amazon.com/eks/latest/userguide/configure-kubectl.html

### Install eksctl via brew
   ```bash
   brew tap weaveworks/tap
   brew install weaveworks/tap/eksctl
   ```
---

## Usage

### Create a cluster

1. Add unique cluster name to conf file `deploy/k8s/eks/simple-cluster.yml`

   ```yml
   ---
   apiVersion: eksctl.io/v1alpha5
   kind: ClusterConfig

   metadata:
     name: <your_cluster_name>
     region: eu-west-1
   ```

2. Run The following command to deploy cluster:
    ```bash
    eksctl create cluster --config-file deploy/k8s/eks/simple-cluster.yml
    ```
    Creation should take around 20 minutes for a simple config.
    eksctl will create your cluster and automatically add the context to your `~/.kube/config`
    The you can proceed with as usual with your tool of choice (kubectl/k9s/lens)

### Delete a cluster

1. Run the delete command below:
   ```
   eksctl delete cluster --region=eu-west-1 --name=<your_cluster_name> --wait
   ```


## Useful references

- [ekcstl config file examples](https://github.com/weaveworks/eksctl/tree/main/examples)  for different deployment scenerios
- [eksctl-docs](https://eksctl.io/introduction/)
- [Aws cli installation](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-version.html)
- [kubectl installation](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/)
