# Rules Status

## K8S CIS Benchmark

### 92/125 implemented rules (74%)

| Rule Number                                          |   Section | Description                                                                                              | Implemented        | Type      |
|------------------------------------------------------|-----------|----------------------------------------------------------------------------------------------------------|--------------------|-----------|
| [1.1.1](bundle/compliance/cis_k8s/rules/cis_1_1_1)   |       1.1 | Ensure that the API server pod specification file permissions are set to 644 or more restrictive         | :white_check_mark: | Automated |
| 1.1.10                                               |       1.1 | Ensure that the Container Network Interface file ownership is set to root:root                           | :x:                | Manual    |
| [1.1.11](bundle/compliance/cis_k8s/rules/cis_1_1_11) |       1.1 | Ensure that the etcd data directory permissions are set to 700 or more restrictive                       | :white_check_mark: | Automated |
| [1.1.12](bundle/compliance/cis_k8s/rules/cis_1_1_12) |       1.1 | Ensure that the etcd data directory ownership is set to etcd:etcd                                        | :white_check_mark: | Automated |
| [1.1.13](bundle/compliance/cis_k8s/rules/cis_1_1_13) |       1.1 | Ensure that the admin.conf file permissions are set to 600                                               | :white_check_mark: | Automated |
| [1.1.14](bundle/compliance/cis_k8s/rules/cis_1_1_14) |       1.1 | Ensure that the admin.conf file ownership is set to root:root                                            | :white_check_mark: | Automated |
| [1.1.15](bundle/compliance/cis_k8s/rules/cis_1_1_15) |       1.1 | Ensure that the scheduler.conf file permissions are set to 644 or more restrictive                       | :white_check_mark: | Automated |
| [1.1.16](bundle/compliance/cis_k8s/rules/cis_1_1_16) |       1.1 | Ensure that the scheduler.conf file ownership is set to root:root                                        | :white_check_mark: | Automated |
| [1.1.17](bundle/compliance/cis_k8s/rules/cis_1_1_17) |       1.1 | Ensure that the controller-manager.conf file permissions are set to 644 or more restrictive              | :white_check_mark: | Automated |
| [1.1.18](bundle/compliance/cis_k8s/rules/cis_1_1_18) |       1.1 | Ensure that the controller-manager.conf file ownership is set to root:root                               | :white_check_mark: | Automated |
| [1.1.19](bundle/compliance/cis_k8s/rules/cis_1_1_19) |       1.1 | Ensure that the Kubernetes PKI directory and file ownership is set to root:root                          | :white_check_mark: | Automated |
| [1.1.2](bundle/compliance/cis_k8s/rules/cis_1_1_2)   |       1.1 | Ensure that the API server pod specification file ownership is set to root:root                          | :white_check_mark: | Automated |
| [1.1.20](bundle/compliance/cis_k8s/rules/cis_1_1_20) |       1.1 | Ensure that the Kubernetes PKI certificate file permissions are set to 644 or more restrictive           | :white_check_mark: | Manual    |
| [1.1.21](bundle/compliance/cis_k8s/rules/cis_1_1_21) |       1.1 | Ensure that the Kubernetes PKI key file permissions are set to 600                                       | :white_check_mark: | Manual    |
| [1.1.3](bundle/compliance/cis_k8s/rules/cis_1_1_3)   |       1.1 | Ensure that the controller manager pod specification file permissions are set to 644 or more restrictive | :white_check_mark: | Automated |
| [1.1.4](bundle/compliance/cis_k8s/rules/cis_1_1_4)   |       1.1 | Ensure that the controller manager pod specification file ownership is set to root:root                  | :white_check_mark: | Automated |
| [1.1.5](bundle/compliance/cis_k8s/rules/cis_1_1_5)   |       1.1 | Ensure that the scheduler pod specification file permissions are set to 644 or more restrictive          | :white_check_mark: | Automated |
| [1.1.6](bundle/compliance/cis_k8s/rules/cis_1_1_6)   |       1.1 | Ensure that the scheduler pod specification file ownership is set to root:root                           | :white_check_mark: | Automated |
| [1.1.7](bundle/compliance/cis_k8s/rules/cis_1_1_7)   |       1.1 | Ensure that the etcd pod specification file permissions are set to 644 or more restrictive               | :white_check_mark: | Automated |
| [1.1.8](bundle/compliance/cis_k8s/rules/cis_1_1_8)   |       1.1 | Ensure that the etcd pod specification file ownership is set to root:root                                | :white_check_mark: | Automated |
| 1.1.9                                                |       1.1 | Ensure that the Container Network Interface file permissions are set to 644 or more restrictive          | :x:                | Manual    |
| 1.2.1                                                |       1.2 | Ensure that the --anonymous-auth argument is set to false                                                | :x:                | Manual    |
| [1.2.10](bundle/compliance/cis_k8s/rules/cis_1_2_10) |       1.2 | Ensure that the admission control plugin EventRateLimit is set                                           | :white_check_mark: | Manual    |
| [1.2.11](bundle/compliance/cis_k8s/rules/cis_1_2_11) |       1.2 | Ensure that the admission control plugin AlwaysAdmit is not set                                          | :white_check_mark: | Automated |
| [1.2.12](bundle/compliance/cis_k8s/rules/cis_1_2_12) |       1.2 | Ensure that the admission control plugin AlwaysPullImages is set                                         | :white_check_mark: | Manual    |
| [1.2.13](bundle/compliance/cis_k8s/rules/cis_1_2_13) |       1.2 | Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used     | :white_check_mark: | Manual    |
| [1.2.14](bundle/compliance/cis_k8s/rules/cis_1_2_14) |       1.2 | Ensure that the admission control plugin ServiceAccount is set                                           | :white_check_mark: | Automated |
| [1.2.15](bundle/compliance/cis_k8s/rules/cis_1_2_15) |       1.2 | Ensure that the admission control plugin NamespaceLifecycle is set                                       | :white_check_mark: | Automated |
| [1.2.16](bundle/compliance/cis_k8s/rules/cis_1_2_16) |       1.2 | Ensure that the admission control plugin NodeRestriction is set                                          | :white_check_mark: | Automated |
| [1.2.17](bundle/compliance/cis_k8s/rules/cis_1_2_17) |       1.2 | Ensure that the --secure-port argument is not set to 0                                                   | :white_check_mark: | Automated |
| [1.2.18](bundle/compliance/cis_k8s/rules/cis_1_2_18) |       1.2 | Ensure that the --profiling argument is set to false                                                     | :white_check_mark: | Automated |
| [1.2.19](bundle/compliance/cis_k8s/rules/cis_1_2_19) |       1.2 | Ensure that the --audit-log-path argument is set                                                         | :white_check_mark: | Automated |
| [1.2.2](bundle/compliance/cis_k8s/rules/cis_1_2_2)   |       1.2 | Ensure that the --token-auth-file parameter is not set                                                   | :white_check_mark: | Automated |
| [1.2.20](bundle/compliance/cis_k8s/rules/cis_1_2_20) |       1.2 | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate                               | :white_check_mark: | Automated |
| [1.2.21](bundle/compliance/cis_k8s/rules/cis_1_2_21) |       1.2 | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate                            | :white_check_mark: | Automated |
| [1.2.22](bundle/compliance/cis_k8s/rules/cis_1_2_22) |       1.2 | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate                             | :white_check_mark: | Automated |
| [1.2.23](bundle/compliance/cis_k8s/rules/cis_1_2_23) |       1.2 | Ensure that the --request-timeout argument is set as appropriate                                         | :white_check_mark: | Manual    |
| [1.2.24](bundle/compliance/cis_k8s/rules/cis_1_2_24) |       1.2 | Ensure that the --service-account-lookup argument is set to true                                         | :white_check_mark: | Automated |
| [1.2.25](bundle/compliance/cis_k8s/rules/cis_1_2_25) |       1.2 | Ensure that the --service-account-key-file argument is set as appropriate                                | :white_check_mark: | Automated |
| [1.2.26](bundle/compliance/cis_k8s/rules/cis_1_2_26) |       1.2 | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate                      | :white_check_mark: | Automated |
| [1.2.27](bundle/compliance/cis_k8s/rules/cis_1_2_27) |       1.2 | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate              | :white_check_mark: | Automated |
| [1.2.28](bundle/compliance/cis_k8s/rules/cis_1_2_28) |       1.2 | Ensure that the --client-ca-file argument is set as appropriate                                          | :white_check_mark: | Automated |
| [1.2.29](bundle/compliance/cis_k8s/rules/cis_1_2_29) |       1.2 | Ensure that the --etcd-cafile argument is set as appropriate                                             | :white_check_mark: | Automated |
| 1.2.3                                                |       1.2 | Ensure that the --DenyServiceExternalIPs is not set                                                      | :x:                | Automated |
| 1.2.30                                               |       1.2 | Ensure that the --encryption-provider-config argument is set as appropriate                              | :x:                | Manual    |
| 1.2.31                                               |       1.2 | Ensure that encryption providers are appropriately configured                                            | :x:                | Manual    |
| [1.2.32](bundle/compliance/cis_k8s/rules/cis_1_2_32) |       1.2 | Ensure that the API Server only makes use of Strong Cryptographic Ciphers                                | :white_check_mark: | Manual    |
| [1.2.4](bundle/compliance/cis_k8s/rules/cis_1_2_4)   |       1.2 | Ensure that the --kubelet-https argument is set to true                                                  | :white_check_mark: | Automated |
| [1.2.5](bundle/compliance/cis_k8s/rules/cis_1_2_5)   |       1.2 | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate   | :white_check_mark: | Automated |
| [1.2.6](bundle/compliance/cis_k8s/rules/cis_1_2_6)   |       1.2 | Ensure that the --kubelet-certificate-authority argument is set as appropriate                           | :white_check_mark: | Automated |
| [1.2.7](bundle/compliance/cis_k8s/rules/cis_1_2_7)   |       1.2 | Ensure that the --authorization-mode argument is not set to AlwaysAllow                                  | :white_check_mark: | Automated |
| [1.2.8](bundle/compliance/cis_k8s/rules/cis_1_2_8)   |       1.2 | Ensure that the --authorization-mode argument includes Node                                              | :white_check_mark: | Automated |
| [1.2.9](bundle/compliance/cis_k8s/rules/cis_1_2_9)   |       1.2 | Ensure that the --authorization-mode argument includes RBAC                                              | :white_check_mark: | Automated |
| 1.3.1                                                |       1.3 | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate                             | :x:                | Manual    |
| [1.3.2](bundle/compliance/cis_k8s/rules/cis_1_3_2)   |       1.3 | Ensure that the --profiling argument is set to false                                                     | :white_check_mark: | Automated |
| [1.3.3](bundle/compliance/cis_k8s/rules/cis_1_3_3)   |       1.3 | Ensure that the --use-service-account-credentials argument is set to true                                | :white_check_mark: | Automated |
| [1.3.4](bundle/compliance/cis_k8s/rules/cis_1_3_4)   |       1.3 | Ensure that the --service-account-private-key-file  argument is set as appropriate                       | :white_check_mark: | Automated |
| [1.3.5](bundle/compliance/cis_k8s/rules/cis_1_3_5)   |       1.3 | Ensure that the --root-ca-file argument is set as appropriate                                            | :white_check_mark: | Automated |
| [1.3.6](bundle/compliance/cis_k8s/rules/cis_1_3_6)   |       1.3 | Ensure that the RotateKubeletServerCertificate argument is set to true                                   | :white_check_mark: | Automated |
| [1.3.7](bundle/compliance/cis_k8s/rules/cis_1_3_7)   |       1.3 | Ensure that the --bind-address argument is set to 127.0.0.1                                              | :white_check_mark: | Automated |
| [1.4.1](bundle/compliance/cis_k8s/rules/cis_1_4_1)   |       1.4 | Ensure that the --profiling argument is set to false                                                     | :white_check_mark: | Automated |
| [1.4.2](bundle/compliance/cis_k8s/rules/cis_1_4_2)   |       1.4 | Ensure that the --bind-address argument is set to 127.0.0.1                                              | :white_check_mark: | Automated |
| [2.1](bundle/compliance/cis_k8s/rules/cis_2_1)       |       2   | Ensure that the --cert-file and --key-file arguments are set as appropriate                              | :white_check_mark: | Automated |
| [2.2](bundle/compliance/cis_k8s/rules/cis_2_2)       |       2   | Ensure that the --client-cert-auth argument is set to true                                               | :white_check_mark: | Automated |
| [2.3](bundle/compliance/cis_k8s/rules/cis_2_3)       |       2   | Ensure that the --auto-tls argument is not set to true                                                   | :white_check_mark: | Automated |
| [2.4](bundle/compliance/cis_k8s/rules/cis_2_4)       |       2   | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate                    | :white_check_mark: | Automated |
| [2.5](bundle/compliance/cis_k8s/rules/cis_2_5)       |       2   | Ensure that the --peer-client-cert-auth argument is set to true                                          | :white_check_mark: | Automated |
| [2.6](bundle/compliance/cis_k8s/rules/cis_2_6)       |       2   | Ensure that the --peer-auto-tls argument is not set to true                                              | :white_check_mark: | Automated |
| 2.7                                                  |       2   | Ensure that a unique Certificate Authority is used for etcd                                              | :x:                | Manual    |
| 3.1.1                                                |       3.1 | Client certificate authentication should not be used for users                                           | :x:                | Manual    |
| 3.2.1                                                |       3.2 | Ensure that a minimal audit policy is created                                                            | :x:                | Manual    |
| 3.2.2                                                |       3.2 | Ensure that the audit policy covers key security concerns                                                | :x:                | Manual    |
| [4.1.1](bundle/compliance/cis_k8s/rules/cis_4_1_1)   |       4.1 | Ensure that the kubelet service file permissions are set to 644 or more restrictive                      | :white_check_mark: | Automated |
| [4.1.10](bundle/compliance/cis_k8s/rules/cis_4_1_10) |       4.1 | Ensure that the kubelet --config configuration file ownership is set to root:root                        | :white_check_mark: | Automated |
| [4.1.2](bundle/compliance/cis_k8s/rules/cis_4_1_2)   |       4.1 | Ensure that the kubelet service file ownership is set to root:root                                       | :white_check_mark: | Automated |
| 4.1.3                                                |       4.1 | If proxy kubeconfig file exists ensure permissions are set to 644 or more restrictive                    | :x:                | Manual    |
| 4.1.4                                                |       4.1 | If proxy kubeconfig file exists ensure ownership is set to root:root                                     | :x:                | Manual    |
| [4.1.5](bundle/compliance/cis_k8s/rules/cis_4_1_5)   |       4.1 | Ensure that the --kubeconfig kubelet.conf file permissions are set to 644 or more restrictive            | :white_check_mark: | Automated |
| [4.1.6](bundle/compliance/cis_k8s/rules/cis_4_1_6)   |       4.1 | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root                             | :white_check_mark: | Automated |
| 4.1.7                                                |       4.1 | Ensure that the certificate authorities file permissions are set to 644 or more restrictive              | :x:                | Manual    |
| 4.1.8                                                |       4.1 | Ensure that the client certificate authorities file ownership is set to root:root                        | :x:                | Manual    |
| [4.1.9](bundle/compliance/cis_k8s/rules/cis_4_1_9)   |       4.1 | Ensure that the kubelet --config configuration file has permissions set to 644 or more restrictive       | :white_check_mark: | Automated |
| [4.2.1](bundle/compliance/cis_k8s/rules/cis_4_2_1)   |       4.2 | Ensure that the --anonymous-auth argument is set to false                                                | :white_check_mark: | Automated |
| [4.2.10](bundle/compliance/cis_k8s/rules/cis_4_2_10) |       4.2 | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate              | :white_check_mark: | Manual    |
| [4.2.11](bundle/compliance/cis_k8s/rules/cis_4_2_11) |       4.2 | Ensure that the --rotate-certificates argument is not set to false                                       | :white_check_mark: | Automated |
| [4.2.12](bundle/compliance/cis_k8s/rules/cis_4_2_12) |       4.2 | Verify that the RotateKubeletServerCertificate argument is set to true                                   | :white_check_mark: | Manual    |
| [4.2.13](bundle/compliance/cis_k8s/rules/cis_4_2_13) |       4.2 | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers                                   | :white_check_mark: | Manual    |
| [4.2.2](bundle/compliance/cis_k8s/rules/cis_4_2_2)   |       4.2 | Ensure that the --authorization-mode argument is not set to AlwaysAllow                                  | :white_check_mark: | Automated |
| [4.2.3](bundle/compliance/cis_k8s/rules/cis_4_2_3)   |       4.2 | Ensure that the --client-ca-file argument is set as appropriate                                          | :white_check_mark: | Automated |
| [4.2.4](bundle/compliance/cis_k8s/rules/cis_4_2_4)   |       4.2 | Verify that the --read-only-port argument is set to 0                                                    | :white_check_mark: | Manual    |
| [4.2.5](bundle/compliance/cis_k8s/rules/cis_4_2_5)   |       4.2 | Ensure that the --streaming-connection-idle-timeout argument is not set to 0                             | :white_check_mark: | Manual    |
| [4.2.6](bundle/compliance/cis_k8s/rules/cis_4_2_6)   |       4.2 | Ensure that the --protect-kernel-defaults argument is set to true                                        | :white_check_mark: | Automated |
| [4.2.7](bundle/compliance/cis_k8s/rules/cis_4_2_7)   |       4.2 | Ensure that the --make-iptables-util-chains argument is set to true                                      | :white_check_mark: | Automated |
| [4.2.8](bundle/compliance/cis_k8s/rules/cis_4_2_8)   |       4.2 | Ensure that the --hostname-override argument is not set                                                  | :white_check_mark: | Manual    |
| [4.2.9](bundle/compliance/cis_k8s/rules/cis_4_2_9)   |       4.2 | Ensure that the --event-qps argument is set to 0 or a level which ensures appropriate event capture      | :white_check_mark: | Manual    |
| 5.1.1                                                |       5.1 | Ensure that the cluster-admin role is only used where required                                           | :x:                | Manual    |
| 5.1.2                                                |       5.1 | Minimize access to secrets                                                                               | :x:                | Manual    |
| [5.1.3](bundle/compliance/cis_k8s/rules/cis_5_1_3)   |       5.1 | Minimize wildcard use in Roles and ClusterRoles                                                          | :white_check_mark: | Manual    |
| 5.1.4                                                |       5.1 | Minimize access to create pods                                                                           | :x:                | Manual    |
| [5.1.5](bundle/compliance/cis_k8s/rules/cis_5_1_5)   |       5.1 | Ensure that default service accounts are not actively used.                                              | :white_check_mark: | Manual    |
| [5.1.6](bundle/compliance/cis_k8s/rules/cis_5_1_6)   |       5.1 | Ensure that Service Account Tokens are only mounted where necessary                                      | :white_check_mark: | Manual    |
| 5.1.7                                                |       5.1 | Avoid use of system:masters group                                                                        | :x:                | Manual    |
| 5.1.8                                                |       5.1 | Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster                    | :x:                | Manual    |
| 5.2.1                                                |       5.2 | Ensure that the cluster has at least one active policy control mechanism in place                        | :x:                | Manual    |
| [5.2.10](bundle/compliance/cis_k8s/rules/cis_5_2_10) |       5.2 | Minimize the admission of containers with capabilities assigned                                          | :white_check_mark: | Manual    |
| 5.2.11                                               |       5.2 | Minimize the admission of Windows HostProcess Containers                                                 | :x:                | Manual    |
| 5.2.12                                               |       5.2 | Minimize the admission of HostPath volumes                                                               | :x:                | Manual    |
| 5.2.13                                               |       5.2 | Minimize the admission of containers which use HostPorts                                                 | :x:                | Manual    |
| [5.2.2](bundle/compliance/cis_k8s/rules/cis_5_2_2)   |       5.2 | Minimize the admission of privileged containers                                                          | :white_check_mark: | Manual    |
| [5.2.3](bundle/compliance/cis_k8s/rules/cis_5_2_3)   |       5.2 | Minimize the admission of containers wishing to share the host process ID namespace                      | :white_check_mark: | Automated |
| [5.2.4](bundle/compliance/cis_k8s/rules/cis_5_2_4)   |       5.2 | Minimize the admission of containers wishing to share the host IPC namespace                             | :white_check_mark: | Automated |
| [5.2.5](bundle/compliance/cis_k8s/rules/cis_5_2_5)   |       5.2 | Minimize the admission of containers wishing to share the host network namespace                         | :white_check_mark: | Automated |
| [5.2.6](bundle/compliance/cis_k8s/rules/cis_5_2_6)   |       5.2 | Minimize the admission of containers with allowPrivilegeEscalation                                       | :white_check_mark: | Automated |
| [5.2.7](bundle/compliance/cis_k8s/rules/cis_5_2_7)   |       5.2 | Minimize the admission of root containers                                                                | :white_check_mark: | Automated |
| [5.2.8](bundle/compliance/cis_k8s/rules/cis_5_2_8)   |       5.2 | Minimize the admission of containers with the NET_RAW capability                                         | :white_check_mark: | Automated |
| [5.2.9](bundle/compliance/cis_k8s/rules/cis_5_2_9)   |       5.2 | Minimize the admission of containers with added capabilities                                             | :white_check_mark: | Automated |
| 5.3.1                                                |       5.3 | Ensure that the CNI in use supports Network Policies                                                     | :x:                | Manual    |
| 5.3.2                                                |       5.3 | Ensure that all Namespaces have Network Policies defined                                                 | :x:                | Manual    |
| 5.4.1                                                |       5.4 | Prefer using secrets as files over secrets as environment variables                                      | :x:                | Manual    |
| 5.4.2                                                |       5.4 | Consider external secret storage                                                                         | :x:                | Manual    |
| 5.5.1                                                |       5.5 | Configure Image Provenance using ImagePolicyWebhook admission controller                                 | :x:                | Manual    |
| 5.7.1                                                |       5.7 | Create administrative boundaries between resources using namespaces                                      | :x:                | Manual    |
| 5.7.2                                                |       5.7 | Ensure that the seccomp profile is set to docker/default in your pod definitions                         | :x:                | Manual    |
| 5.7.3                                                |       5.7 | Apply Security Context to Your Pods and Containers                                                       | :x:                | Manual    |
| 5.7.4                                                |       5.7 | The default namespace should not be used                                                                 | :x:                | Manual    |

## EKS CIS Benchmark

### 31/52 implemented rules (60%)

| Rule Number                                          |   Section | Description                                                                                              | Implemented        | Type      |
|------------------------------------------------------|-----------|----------------------------------------------------------------------------------------------------------|--------------------|-----------|
| [2.1.1](bundle/compliance/cis_eks/rules/cis_2_1_1)   |       2.1 | Enable audit Logs                                                                                        | :white_check_mark: | Manual    |
| [3.1.1](bundle/compliance/cis_eks/rules/cis_3_1_1)   |       3.1 | Ensure that the kubeconfig file permissions are set to 644 or more restrictive                           | :white_check_mark: | Manual    |
| [3.1.2](bundle/compliance/cis_eks/rules/cis_3_1_2)   |       3.1 | Ensure that the kubelet kubeconfig file ownership is set to root:root                                    | :white_check_mark: | Manual    |
| [3.1.3](bundle/compliance/cis_eks/rules/cis_3_1_3)   |       3.1 | Ensure that the kubelet configuration file has permissions set to 644 or more restrictive                | :white_check_mark: | Manual    |
| [3.1.4](bundle/compliance/cis_eks/rules/cis_3_1_4)   |       3.1 | Ensure that the kubelet configuration file ownership is set to root:root                                 | :white_check_mark: | Manual    |
| [3.2.1](bundle/compliance/cis_eks/rules/cis_3_2_1)   |       3.2 | Ensure that the --anonymous-auth argument is set to false                                                | :white_check_mark: | Automated |
| [3.2.10](bundle/compliance/cis_eks/rules/cis_3_2_10) |       3.2 | Ensure that the --rotate-certificates argument is not set to false                                       | :white_check_mark: | Manual    |
| [3.2.11](bundle/compliance/cis_eks/rules/cis_3_2_11) |       3.2 | Ensure that the RotateKubeletServerCertificate argument is set to true                                   | :white_check_mark: | Manual    |
| [3.2.2](bundle/compliance/cis_eks/rules/cis_3_2_2)   |       3.2 | Ensure that the --authorization-mode argument is not set to AlwaysAllow                                  | :white_check_mark: | Automated |
| [3.2.3](bundle/compliance/cis_eks/rules/cis_3_2_3)   |       3.2 | Ensure that the --client-ca-file argument is set as appropriate                                          | :white_check_mark: | Manual    |
| [3.2.4](bundle/compliance/cis_eks/rules/cis_3_2_4)   |       3.2 | Ensure that the --read-only-port is secured                                                              | :white_check_mark: | Manual    |
| [3.2.5](bundle/compliance/cis_eks/rules/cis_3_2_5)   |       3.2 | Ensure that the --streaming-connection-idle-timeout argument is not set to 0                             | :white_check_mark: | Manual    |
| [3.2.6](bundle/compliance/cis_eks/rules/cis_3_2_6)   |       3.2 | Ensure that the --protect-kernel-defaults argument is set to true                                        | :white_check_mark: | Automated |
| [3.2.7](bundle/compliance/cis_eks/rules/cis_3_2_7)   |       3.2 | Ensure that the --make-iptables-util-chains argument is set to true                                      | :white_check_mark: | Automated |
| [3.2.8](bundle/compliance/cis_eks/rules/cis_3_2_8)   |       3.2 | Ensure that the --hostname-override argument is not set                                                  | :white_check_mark: | Manual    |
| [3.2.9](bundle/compliance/cis_eks/rules/cis_3_2_9)   |       3.2 | Ensure that the --eventRecordQPS argument is set to 0 or a level which ensures appropriate event capture | :white_check_mark: | Automated |
| 4.1.1                                                |       4.1 | Ensure that the cluster-admin role is only used where required                                           | :x:                | Manual    |
| 4.1.2                                                |       4.1 | Minimize access to secrets                                                                               | :x:                | Manual    |
| 4.1.3                                                |       4.1 | Minimize wildcard use in Roles and ClusterRoles                                                          | :x:                | Manual    |
| 4.1.4                                                |       4.1 | Minimize access to create pods                                                                           | :x:                | Manual    |
| 4.1.5                                                |       4.1 | Ensure that default service accounts are not actively used.                                              | :x:                | Manual    |
| 4.1.6                                                |       4.1 | Ensure that Service Account Tokens are only mounted where necessary                                      | :x:                | Manual    |
| [4.2.1](bundle/compliance/cis_eks/rules/cis_4_2_1)   |       4.2 | Minimize the admission of privileged containers                                                          | :white_check_mark: | Automated |
| [4.2.2](bundle/compliance/cis_eks/rules/cis_4_2_2)   |       4.2 | Minimize the admission of containers wishing to share the host process ID namespace                      | :white_check_mark: | Automated |
| [4.2.3](bundle/compliance/cis_eks/rules/cis_4_2_3)   |       4.2 | Minimize the admission of containers wishing to share the host IPC namespace                             | :white_check_mark: | Automated |
| [4.2.4](bundle/compliance/cis_eks/rules/cis_4_2_4)   |       4.2 | Minimize the admission of containers wishing to share the host network namespace                         | :white_check_mark: | Automated |
| [4.2.5](bundle/compliance/cis_eks/rules/cis_4_2_5)   |       4.2 | Minimize the admission of containers with allowPrivilegeEscalation                                       | :white_check_mark: | Automated |
| [4.2.6](bundle/compliance/cis_eks/rules/cis_4_2_6)   |       4.2 | Minimize the admission of root containers                                                                | :white_check_mark: | Automated |
| [4.2.7](bundle/compliance/cis_eks/rules/cis_4_2_7)   |       4.2 | Minimize the admission of containers with the NET_RAW capability                                         | :white_check_mark: | Automated |
| [4.2.8](bundle/compliance/cis_eks/rules/cis_4_2_8)   |       4.2 | Minimize the admission of containers with added capabilities                                             | :white_check_mark: | Automated |
| [4.2.9](bundle/compliance/cis_eks/rules/cis_4_2_9)   |       4.2 | Minimize the admission of containers with capabilities assigned                                          | :white_check_mark: | Manual    |
| 4.3.1                                                |       4.3 | Ensure latest CNI version is used                                                                        | :x:                | Manual    |
| 4.3.2                                                |       4.3 | Ensure that all Namespaces have Network Policies defined                                                 | :x:                | Automated |
| 4.4.1                                                |       4.4 | Prefer using secrets as files over secrets as environment variables                                      | :x:                | Manual    |
| 4.4.2                                                |       4.4 | Consider external secret storage                                                                         | :x:                | Manual    |
| 4.5.1                                                |       4.5 | Configure Image Provenance using ImagePolicyWebhook admission controller                                 | :x:                | Manual    |
| 4.6.1                                                |       4.6 | Create administrative boundaries between resources using namespaces                                      | :x:                | Manual    |
| 4.6.2                                                |       4.6 | Apply Security Context to Your Pods and Containers                                                       | :x:                | Manual    |
| 4.6.3                                                |       4.6 | The default namespace should not be used                                                                 | :x:                | Automated |
| [5.1.1](bundle/compliance/cis_eks/rules/cis_5_1_1)   |       5.1 | Ensure Image Vulnerability Scanning using Amazon ECR image scanning or a third party provider            | :white_check_mark: | Manual    |
| 5.1.2                                                |       5.1 | Minimize user access to Amazon ECR                                                                       | :x:                | Manual    |
| 5.1.3                                                |       5.1 | Minimize cluster access to read-only for Amazon ECR                                                      | :x:                | Manual    |
| 5.1.4                                                |       5.1 | Minimize Container Registries to only those approved                                                     | :x:                | Manual    |
| 5.2.1                                                |       5.2 | Prefer using dedicated EKS Service Accounts                                                              | :x:                | Manual    |
| [5.3.1](bundle/compliance/cis_eks/rules/cis_5_3_1)   |       5.3 | Ensure Kubernetes Secrets are encrypted using Customer Master Keys (CMKs) managed in AWS KMS             | :white_check_mark: | Automated |
| [5.4.1](bundle/compliance/cis_eks/rules/cis_5_4_1)   |       5.4 | Restrict Access to the Control Plane Endpoint                                                            | :white_check_mark: | Manual    |
| [5.4.2](bundle/compliance/cis_eks/rules/cis_5_4_2)   |       5.4 | Ensure clusters are created with Private Endpoint Enabled and Public Access Disabled                     | :white_check_mark: | Manual    |
| [5.4.3](bundle/compliance/cis_eks/rules/cis_5_4_3)   |       5.4 | Ensure clusters are created with Private Nodes                                                           | :white_check_mark: | Manual    |
| 5.4.4                                                |       5.4 | Ensure Network Policy is Enabled and set as appropriate                                                  | :x:                | Manual    |
| [5.4.5](bundle/compliance/cis_eks/rules/cis_5_4_5)   |       5.4 | Encrypt traffic to HTTPS load balancers with TLS certificates                                            | :white_check_mark: | Manual    |
| 5.5.1                                                |       5.5 | Manage Kubernetes RBAC users with AWS IAM Authenticator for Kubernetes                                   | :x:                | Manual    |
| 5.6.1                                                |       5.6 | Consider Fargate for running untrusted workloads                                                         | :x:                | Manual    |

## AWS CIS Benchmark

### 2/63 implemented rules (3%)

| Rule Number                                    |   Section | Description                                                                                                        | Implemented        | Type      |
|------------------------------------------------|-----------|--------------------------------------------------------------------------------------------------------------------|--------------------|-----------|
| 1.1                                            |       1   | Maintain current contact details                                                                                   | :x:                | Manual    |
| 1.10                                           |       1   | Ensure multi-factor authentication (MFA) is enabled for all IAM users that have a console password                 | :x:                | Automated |
| 1.11                                           |       1   | Do not setup access keys during initial user setup for all IAM users that have a console password                  | :x:                | Automated |
| 1.12                                           |       1   | Ensure credentials unused for 45 days or greater are disabled                                                      | :x:                | Automated |
| 1.13                                           |       1   | Ensure there is only one active access key available for any single IAM user                                       | :x:                | Automated |
| 1.14                                           |       1   | Ensure access keys are rotated every 90 days or less                                                               | :x:                | Automated |
| 1.15                                           |       1   | Ensure IAM Users Receive Permissions Only Through Groups                                                           | :x:                | Automated |
| 1.16                                           |       1   | Ensure IAM policies that allow full "*:*" administrative privileges are not attached                               | :x:                | Automated |
| 1.17                                           |       1   | Ensure a support role has been created to manage incidents with AWS Support                                        | :x:                | Automated |
| 1.18                                           |       1   | Ensure IAM instance roles are used for AWS resource access from instances                                          | :x:                | Manual    |
| 1.19                                           |       1   | Ensure that all the expired SSL/TLS certificates stored in AWS IAM are removed                                     | :x:                | Automated |
| 1.2                                            |       1   | Ensure security contact information is registered                                                                  | :x:                | Manual    |
| 1.20                                           |       1   | Ensure that IAM Access analyzer is enabled for all regions                                                         | :x:                | Automated |
| 1.21                                           |       1   | Ensure IAM users are managed centrally via identity federation or AWS Organizations for multi-account environments | :x:                | Manual    |
| 1.3                                            |       1   | Ensure security questions are registered in the AWS account                                                        | :x:                | Manual    |
| 1.4                                            |       1   | Ensure no 'root' user account access key exists                                                                    | :x:                | Automated |
| 1.5                                            |       1   | Ensure MFA is enabled for the 'root' user account                                                                  | :x:                | Automated |
| 1.6                                            |       1   | Ensure hardware MFA is enabled for the 'root' user account                                                         | :x:                | Automated |
| 1.7                                            |       1   | Eliminate use of the 'root' user for administrative and daily tasks                                                | :x:                | Automated |
| [1.8](bundle/compliance/cis_aws/rules/cis_1_8) |       1   | Ensure IAM password policy requires minimum length of 14 or greater                                                | :white_check_mark: | Automated |
| [1.9](bundle/compliance/cis_aws/rules/cis_1_9) |       1   | Ensure IAM password policy prevents password reuse                                                                 | :white_check_mark: | Automated |
| 2.1.1                                          |       2.1 | Ensure all S3 buckets employ encryption-at-rest                                                                    | :x:                | Automated |
| 2.1.2                                          |       2.1 | Ensure S3 Bucket Policy is set to deny HTTP requests                                                               | :x:                | Automated |
| 2.1.3                                          |       2.1 | Ensure MFA Delete is enabled on S3 buckets                                                                         | :x:                | Automated |
| 2.1.4                                          |       2.1 | Ensure all data in Amazon S3 has been discovered, classified and secured when required.                            | :x:                | Manual    |
| 2.1.5                                          |       2.1 | Ensure that S3 Buckets are configured with 'Block public access (bucket settings)'                                 | :x:                | Automated |
| 2.2.1                                          |       2.2 | Ensure EBS Volume Encryption is Enabled in all Regions                                                             | :x:                | Automated |
| 2.3.1                                          |       2.3 | Ensure that encryption is enabled for RDS Instances                                                                | :x:                | Automated |
| 2.3.2                                          |       2.3 | Ensure Auto Minor Version Upgrade feature is Enabled for RDS Instances                                             | :x:                | Automated |
| 2.3.3                                          |       2.3 | Ensure that public access is not given to RDS Instance                                                             | :x:                | Automated |
| 2.4.1                                          |       2.4 | Ensure that encryption is enabled for EFS file systems                                                             | :x:                | Manual    |
| 3.1                                            |       3   | Ensure CloudTrail is enabled in all regions                                                                        | :x:                | Automated |
| 3.10                                           |       3   | Ensure that Object-level logging for write events is enabled for S3 bucket                                         | :x:                | Automated |
| 3.11                                           |       3   | Ensure that Object-level logging for read events is enabled for S3 bucket                                          | :x:                | Automated |
| 3.2                                            |       3   | Ensure CloudTrail log file validation is enabled                                                                   | :x:                | Automated |
| 3.3                                            |       3   | Ensure the S3 bucket used to store CloudTrail logs is not publicly accessible                                      | :x:                | Automated |
| 3.4                                            |       3   | Ensure CloudTrail trails are integrated with CloudWatch Logs                                                       | :x:                | Automated |
| 3.5                                            |       3   | Ensure AWS Config is enabled in all regions                                                                        | :x:                | Automated |
| 3.6                                            |       3   | Ensure S3 bucket access logging is enabled on the CloudTrail S3 bucket                                             | :x:                | Automated |
| 3.7                                            |       3   | Ensure CloudTrail logs are encrypted at rest using KMS CMKs                                                        | :x:                | Automated |
| 3.8                                            |       3   | Ensure rotation for customer created symmetric CMKs is enabled                                                     | :x:                | Automated |
| 3.9                                            |       3   | Ensure VPC flow logging is enabled in all VPCs                                                                     | :x:                | Automated |
| 4.1                                            |       4   | Ensure a log metric filter and alarm exist for unauthorized API calls                                              | :x:                | Automated |
| 4.10                                           |       4   | Ensure a log metric filter and alarm exist for security group changes                                              | :x:                | Automated |
| 4.11                                           |       4   | Ensure a log metric filter and alarm exist for changes to Network Access Control Lists (NACL)                      | :x:                | Automated |
| 4.12                                           |       4   | Ensure a log metric filter and alarm exist for changes to network gateways                                         | :x:                | Automated |
| 4.13                                           |       4   | Ensure a log metric filter and alarm exist for route table changes                                                 | :x:                | Automated |
| 4.14                                           |       4   | Ensure a log metric filter and alarm exist for VPC changes                                                         | :x:                | Automated |
| 4.15                                           |       4   | Ensure a log metric filter and alarm exists for AWS Organizations changes                                          | :x:                | Automated |
| 4.16                                           |       4   | Ensure AWS Security Hub is enabled                                                                                 | :x:                | Automated |
| 4.2                                            |       4   | Ensure a log metric filter and alarm exist for Management Console sign-in without MFA                              | :x:                | Automated |
| 4.3                                            |       4   | Ensure a log metric filter and alarm exist for usage of 'root' account                                             | :x:                | Automated |
| 4.4                                            |       4   | Ensure a log metric filter and alarm exist for IAM policy changes                                                  | :x:                | Automated |
| 4.5                                            |       4   | Ensure a log metric filter and alarm exist for CloudTrail configuration changes                                    | :x:                | Automated |
| 4.6                                            |       4   | Ensure a log metric filter and alarm exist for AWS Management Console authentication failures                      | :x:                | Automated |
| 4.7                                            |       4   | Ensure a log metric filter and alarm exist for disabling or scheduled deletion of customer created CMKs            | :x:                | Automated |
| 4.8                                            |       4   | Ensure a log metric filter and alarm exist for S3 bucket policy changes                                            | :x:                | Automated |
| 4.9                                            |       4   | Ensure a log metric filter and alarm exist for AWS Config configuration changes                                    | :x:                | Automated |
| 5.1                                            |       5   | Ensure no Network ACLs allow ingress from 0.0.0.0/0 to remote server administration ports                          | :x:                | Automated |
| 5.2                                            |       5   | Ensure no security groups allow ingress from 0.0.0.0/0 to remote server administration ports                       | :x:                | Automated |
| 5.3                                            |       5   | Ensure no security groups allow ingress from ::/0 to remote server administration ports                            | :x:                | Automated |
| 5.4                                            |       5   | Ensure the default security group of every VPC restricts all traffic                                               | :x:                | Automated |
| 5.5                                            |       5   | Ensure routing tables for VPC peering are "least access"                                                           | :x:                | Manual    |
