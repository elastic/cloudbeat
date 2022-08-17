# CIS Policies Generator
Running the generator will create/modify a metadata file for each implemented rule of the provided benchmarks in the `input` folder.

`npm start -- --help`

```
Usage: cis-policies-generator [options]

CIS Benchmark parser CLI

Options:
  -V, --version             output the version number
  -t, --templates           generate csp rule templates and place them in the integration dir
  -b, --benchmark <string>  benchmark to be used for the rules template generation (default: "cis_k8s")
  -m, --rulesMeta           generate rules metadata for any provided benchmark in the input dir
  -h, --help                display help for command

    Example calls:
        $ npm start -- -m
        $ npm start -- -t -b 'cis_eks'
```


Example serialization of a single rule:

```
metadata:
  id: 883ab83b-8dbc-5072-aef7-0f4c4a7f4048
  name: Ensure that the API server pod specification file permissions are set to
    644 or more restrictive
  profile_applicability: "* Level 1 - Master Node"
  description: Ensure that the API server pod specification file has permissions
    of `644` or more restrictive.
  rationale: The API server pod specification file controls various parameters
    that set the behavior of the API server. You should restrict its file
    permissions to maintain the integrity of the file. The file should be
    writable by only the administrators on the system.
  audit: >-
    Run the below command (based on the file location on your system) on the
    Control Plane node. For example,


    ```

    stat -c %a /etc/kubernetes/manifests/kube-apiserver.yaml

    ```


    Verify that the permissions are `644` or more restrictive.
  remediation: >
    Run the below command (based on the file location on your system)
    on the Control Plane node. For example,


    ```

    chmod 644 /etc/kubernetes/manifests/kube-apiserver.yaml

    ```
  impact: None
  references:
    - https://kubernetes.io/docs/admin/kube-apiserver/
  tags:
    - CIS
    - Kubernetes
    - CIS 1.1.1
    - Control Plane Node Configuration Files
  section: Control Plane Node Configuration Files
  benchmark:
    name: CIS Kubernetes V1.23
    version: v1.0.1
    id: cis_k8s_v1.23
  default_value: |
    By default, the `kube-apiserver.yaml` file has permissions of `640`.

```
