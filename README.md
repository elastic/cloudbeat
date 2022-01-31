![Coverage Badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/oren-zohar/a7160df46e48dff45b24096de9302d38/raw/csp-security-policies_coverage.json)

# Cloud Security Posture - Rego policies
    .
    ├── compliance                         # Compliance policies
    │   ├── lib
    │   │   ├── common.rego                # Common functions
    │   │   ├── common_tests.rego          # Common functions tests
    │   │   ├── data_adapter.rego          # Input data adapter
    │   │   └── test.rego                  # Common Test functions
    │   ├── cis_k8s
    │   │   ├── cis_k8s.rego               # Handles all Kubernetes CIS rules evalutations
    │   │   ├── test_data.rego             # CIS Test data generators
    │   │   ├── schemas                    # Benchmark's schemas
    │   │   │   └── input_scehma.rego
    │   │   ├── rules
    │   │   │   ├── cis_1_1_1              # CIS 1.1.1 rule package
    │   │   │   │   ├── rule.rego
    │   │   │   │   └── test.rego
    │   │   │   └── ...
    └── main.rego                          # Evaluates all policies and returns the findings

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
    "type": "file-system",
    "resource": {
        "mode": "0700",
        "path": "/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml",
        "uid": "etc",
        "filename": "kube-apiserver.yaml",
        "gid": "root"
    }
}
```

### Evaluate entire policy into output.json
```console
opa eval data.main --format pretty -i input.json -b . > output.json
```

### Evaluate findings only
```console
opa eval data.main.findings --format pretty -i input.json -b . > output.json
```

<details>
<summary>Example output</summary>

```json
{
  "findings": [
    {
      "result": {
        "evaluation": "failed",
        "evidence": {
          "filemode": "0700"
        }
      },
      "rule": {
        "benchmark": "CIS Kubernetes",
        "description": "The API server pod specification file controls various parameters that set the behavior of the API server. You should restrict its file permissions to maintain the integrity of the file. The file should be writable by only the administrators on the system.",
        "impact": "None",
        "name": "Ensure that the API server pod specification file permissions are set to 644 or more restrictive",
        "remediation": "chmod 644 /etc/kubernetes/manifests/kube-apiserver.yaml",
        "tags": [
          "CIS",
          "CIS v1.6.0",
          "Kubernetes",
          "CIS 1.1.1",
          "Master Node Configuration"
        ]
      }
    },
    {
      "result": {
        "evaluation": "passed",
        "evidence": {
          "gid": "root",
          "uid": "root"
        }
      },
      "rule": {
        "benchmark": "CIS Kubernetes",
        "description": "The API server pod specification file controls various parameters that set the behavior of the API server. You should set its file ownership to maintain the integrity of the file. The file should be owned by root:root.",
        "impact": "None",
        "name": "Ensure that the API server pod specification file ownership is set to root:root",
        "remediation": "chown root:root /etc/kubernetes/manifests/kube-apiserver.yaml",
        "tags": [
          "CIS",
          "CIS v1.6.0",
          "Kubernetes",
          "CIS 1.1.2",
          "Master Node Configuration"
        ]
      }
    }
  ],
  "resource": {
    "filename": "kube-apiserver.yaml",
    "gid": "root",
    "mode": "0700",
    "path": "/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml",
    "type": "file-system",
    "uid": "root"
  }
}
```

</details>

### Evaluate with input schema

```console
❯ opa eval data.main --format pretty -i input.json -b . -s compliance/cis_k8s/schemas/input_schema.json
1 error occurred: compliance/lib/data_adapter.rego:11: rego_type_error: undefined ref: input.filenames
        input.filenames
              ^
              have: "filenames"
              want (one of): ["command" "filename" "gid" "mode" "path" "type" "uid"]

```
## Local Testing
### Test entire policy
```console
opa test -v compliance
```

### Test specific rule
```console
opa test -v compliance/lib compliance/cis_k8s/test_data.rego compliance/cis_k8s/rules/cis_1_1_2 --ignore="common_tests.rego"
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
            "type": "file-system",
            "mode": "0700",
            "path": "/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml",
            "uid": "etc",
            "filename": "kube-apiserver.yaml",
            "gid": "root"
        }
    }
}'
```