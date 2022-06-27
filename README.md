![Coverage Badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/oren-zohar/a7160df46e48dff45b24096de9302d38/raw/csp-security-policies_coverage.json)

# Cloud Security Posture - Rego policies

    .
    ├── README.md
    ├── bundle
    │   ├── builder.go                            # Bundle building code
    │   ├── compliance                            # Compliance policies
    │   │   ├── cis_eks
    │   │   │   ├── cis_eks.rego                  # Handles all EKS CIS rules evalutations
    │   │   │   ├── data_adapter.rego
    │   │   │   ├── rules
    │   │   │   │   ├── cis_2_1_1                 # CIS EKS 2.1.1 rule package
    │   │   │   │   │   ├── data.yaml             # Rule's metadata
    │   │   │   │   │   ├── rule.rego             # Rule's rego
    │   │   │   │   │   └── test.rego             # Rule's test
    |   |   |   |   ├── ...
    │   │   │   └── test_data.rego                # CIS EKS Test data generators
    │   │   ├── cis_k8s
    │   │   │   ├── cis_k8s.rego                  # Handles all Kubernetes CIS rules evalutations
    │   │   │   ├── data_adapter.rego
    │   │   │   ├── rules
    │   │   │   │   ├── cis_1_1_1
    │   │   │   │   │   ├── data.yaml
    │   │   │   │   │   ├── rule.rego
    │   │   │   │   │   └── test.rego
    |   |   |   |   ├── ...
    │   │   │   └── schemas                       # Benchmark's schemas
    │   │   │   └── input_schema.json
    │   │   ├── kubernetes_common
    │   │   │   └── test_data.rego
    │   │   ├── lib
    │   │   │   ├── assert.rego
    │   │   │   ├── common                        # Common functions and tests
    │   │   │   │   ├── common.rego
    │   │   │   │   └── test.rego
    │   │   │   ├── data_adapter                  # Input data adapter and tests
    │   │   │   │   ├── data_adapter.rego
    │   │   │   │   └── test.rego
    │   │   │   ├── output_validations            # Output validations for tests
    │   │   │   │   ├── output_validations.rego
    │   │   │   │   └── test.rego
    │   │   │   └── test.rego
    │   │   └── main.rego                         # Evaluates all policies and returns the findings
    │   ├── embed.go                              # Embed of benchmarks
    │   ├── server.go                             # Hosting and creation of bundle server functions
    │   └── server_test.go
    ├── main.go
    └── server
    └── host.go                                   # Hosting and creation of bundle server for benchmarks

## Local Evaluation

Add the following configuration files into the root folder

##### `data.yaml`

should contain the list of rules you want to evaluate (also supports json)

```yaml
activated_rules:
  cis_k8s:
    - cis_1_1_1
    - cis_1_1_2
  cis_eks:
    - cis_3_1_1
    - cis_3_1_2
```

##### `input.json`

should contain an beat/agent output, e.g. filesystem data

```json
{
  "type": "file",
  "resource": {
    "mode": "0700",
    "path": "/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml",
    "owner": "etc",
    "group": "root",
    "name": "kube-apiserver.yaml",
    "gid": 20,
    "uid": 501
  }
}
```

### Evaluate entire policy into output.json

```console
opa eval data.main --format pretty -i input.json -b ./bundle > output.json
```

### Evaluate findings only

```console
opa eval data.main.findings --format pretty -i input.json -b ./bundle > output.json
```

<details>
<summary>Example output</summary>

````json
{
  "findings": [
    {
      "result": {
        "evaluation": "failed",
        "expected": {
          "filemode": "0644"
        },
        "evidence": {
          "filemode": "0700"
        }
      },
      "rule": {
        "id": "59b5a77b-b090-5630-9a33-73eb805b2d52",
        "name": "Ensure that the API server pod specification file permissions are set to 644 or more restrictive (Automated)",
        "profile_applicability": "* Level 1 - Master Node\n",
        "description": "Ensure that the API server pod specification file has permissions of `644` or more restrictive.\n",
        "rationale": "The API server pod specification file controls various parameters that set the behavior of the API server. You should restrict its file permissions to maintain the integrity of the file. The file should be writable by only the administrators on the system.\n",
        "audit": "Run the below command (based on the file location on your system) on the\nmaster node.\nFor example,\n```\nstat -c %a /etc/kubernetes/manifests/kube-apiserver.yaml\n```\nVerify that the permissions are `644` or more restrictive.\n",
        "remediation": "Run the below command (based on the file location on your system) on the\nmaster node.\nFor example,\n```\nchmod 644 /etc/kubernetes/manifests/kube-apiserver.yaml\n```\n",
        "impact": "None\n",
        "default_value": "By default, the `kube-apiserver.yaml` file has permissions of `640`.\n",
        "references": "1. [https://kubernetes.io/docs/admin/kube-apiserver/](https://kubernetes.io/docs/admin/kube-apiserver/)\n",
        "section": "Master Node Configuration Files",
        "version": 1,
        "tags": [
          "CIS",
          "Kubernetes",
          "CIS 1.1.1",
          "Master Node Configuration Files"
        ],
        "benchmark": {
          "name": "CIS Kubernetes V1.20",
          "version": "v1.0.0"
        }
      }
    },
    {
      "result": {
        "evaluation": "passed",
        "expected": {
          "group": "root",
          "owner": "root"
        },
        "evidence": {
          "group": "root",
          "owner": "root"
        }
      },
      "rule": {
        "id": "9f318d4d-2451-574a-99dc-838ed213f09b",
        "name": "Ensure that the API server pod specification file ownership is set toroot:root (Automated)",
        "profile_applicability": "* Level 1 - Master Node\n",
        "description": "Ensure that the API server pod specification file ownership is set to `root:root`.\n",
        "rationale": "The API server pod specification file controls various parameters that set the behavior of the API server. You should set its file ownership to maintain the integrity of the file. The file should be owned by `root:root`.\n",
        "audit": "Run the below command (based on the file location on your system) on the\nmaster node.\nFor example,\n```\nstat -c %U:%G /etc/kubernetes/manifests/kube-apiserver.yaml\n```\nVerify that the ownership is set to `root:root`.\n",
        "remediation": "Run the below command (based on the file location on your system) on the\nmaster node.\nFor example,\n```\nchown root:root /etc/kubernetes/manifests/kube-apiserver.yaml\n```\n",
        "impact": "None\n",
        "default_value": "By default, the `kube-apiserver.yaml` file ownership is set to `root:root`.\n",
        "references": "1. [https://kubernetes.io/docs/admin/kube-apiserver/](https://kubernetes.io/docs/admin/kube-apiserver/)\n",
        "section": "Master Node Configuration Files",
        "version": 1,
        "tags": [
          "CIS",
          "Kubernetes",
          "CIS 1.1.2",
          "Master Node Configuration Files"
        ],
        "benchmark": {
          "name": "CIS Kubernetes V1.20",
          "version": "v1.0.0"
        }
      }
    }
  ],
  "resource": {
    "name": "kube-apiserver.yaml",
    "group": "root",
    "mode": "0700",
    "path": "/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml",
    "type": "file",
    "owner": "root",
    "uid": 501,
    "gid": 20
  }
}
````

</details>

### Evaluate with input schema

```console
❯ opa eval data.main --format pretty -i input.json -b ./bundle -s bundle/compliance/cis_k8s/schemas/input_schema.json
1 error occurred: bundle/compliance/lib/data_adapter.rego:11: rego_type_error: undefined ref: input.filenames
        input.filenames
              ^
              have: "filenames"
              want (one of): ["command" "filename" "gid" "mode" "path" "type" "uid"]

```

## Local Testing

### Test entire policy

```console
opa build -b ./bundle -e ./bundle/compliance
```

```console
opa test -b bundle.tar.gz -v
```

### Test specific rule

```console
opa test -v bundle/compliance/kubernetes_common bundle/compliance/lib bundle/compliance/cis_k8s/test_data.rego bundle/compliance/cis_k8s/rules/cis_1_1_2 --ignore="common_tests.rego"
```

### Pre-commit hooks

see [pre-commit](https://pre-commit.com/) package

- Install the package `brew install pre-commit`
- Then run `pre-commit install`
- Finally `pre-commit run --all-files --verbose`

### Running opa server with the compliance policy

```console
docker run --rm -p 8181:8181 -v $(pwd):/bundle openpolicyagent/opa:0.36.1 run -s -b /bundle
```

Test it 🚀

```curl
curl --location --request POST 'http://localhost:8181/v1/data/main' \
--header 'Content-Type: application/json' \
--data-raw '{
    "input": {
        "resource": {
            "type": "file",
            "mode": "0700",
            "path": "/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml",
            "uid": "etc",
            "name": "kube-apiserver.yaml",
            "group": "root"
        }
    }
}'
```
