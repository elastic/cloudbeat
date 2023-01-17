# Cloud Security Posture - Rego policies

[![CIS K8S](https://img.shields.io/badge/CIS-Kubernetes%20(74%25)-326CE5?logo=Kubernetes)](RULES.md#k8s-cis-benchmark)
[![CIS EKS](https://img.shields.io/badge/CIS-Amazon%20EKS%20(60%25)-FF9900?logo=Amazon+EKS)](RULES.md#eks-cis-benchmark)
[![CIS AWS](https://img.shields.io/badge/CIS-AWS%20(30%25)-232F3E?logo=Amazon+AWS)](RULES.md#aws-cis-benchmark)

![Coverage Badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/oren-zohar/a7160df46e48dff45b24096de9302d38/raw/csp-security-policies_coverage.json)

<details>
<summary>Project structure</summary>

    .
    ├── bundle
    │   ├── compliance                         # Compliance policies
    │   │   ├── cis_aws
    │   │   │   ├── rules
    │   │   │   │   ├── cis_1_8                # CIS AWS 1.8 rule package
    │   │   │   │   │   ├── data.yaml          # Rule's metadata
    │   │   │   │   │   ├── rule.rego          # Rule's rego
    │   │   │   │   │   └── test.rego          # Rule's test
    │   │   │   │   ...
    │   │   ├── cis_eks
    │   │   │   ├── rules
    │   │   ├── cis_k8s
    │   │   │   ├── rules
    │   │   │   ├── schemas                    # Benchmark's schemas
    │   │   ├── kubernetes_common
    │   │   ├── lib
    │   │   │   ├── common                     # Common functions and tests
    │   │   │   ├── output_validations
    │   │   ├── policy                         # Common audit functions per input
    │   │   │   ├── kube_api
    │   │   │   ...
    ├── dev
    └── server

</details>

## Local Evaluation

**`input.json`**

should contain a beat/agent output and the `benchmark` (not mandatory - without specifying benchmark all benchmarks will
apply), e.g. k8s eks aws

```json
{
  "type": "file",
  "benchmark": "cis_k8s",
  "sub_type": "file",
  "resource": {
    "mode": "700",
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

```bash
opa eval data.main --format pretty -i input.json -b ./bundle > output.json
```

### Evaluate findings only

```bash
opa eval data.main.findings --format pretty -i input.json -b ./bundle > output.json
```

<details>
<summary>Example output</summary>

````json
{
  "result": {
    "evaluation": "failed",
    "evidence": {
      "filemode": "700"
    },
    "expected": {
      "filemode": "644"
    }
  },
  "rule": {
    "audit": "Run the below command (based on the file location on your system) on the\ncontrol plane node.\nFor example,\n```\nstat -c %a /etc/kubernetes/manifests/kube-apiserver.yaml\n```\nVerify that the permissions are `644` or more restrictive.\n",
    "benchmark": {
      "id": "cis_k8s",
      "name": "CIS Kubernetes V1.23",
      "version": "v1.0.0"
    },
    "default_value": "By default, the `kube-apiserver.yaml` file has permissions of `640`.\n",
    "description": "Ensure that the API server pod specification file has permissions of `644` or more restrictive.\n",
    "id": "6664c1b8-05f2-5872-a516-4b2c3c36d2d7",
    "impact": "None\n",
    "name": "Ensure that the API server pod specification file permissions are set to 644 or more restrictive (Automated)",
    "profile_applicability": "* Level 1 - Master Node\n",
    "rationale": "The API server pod specification file controls various parameters that set the behavior of the API server. You should restrict its file permissions to maintain the integrity of the file. The file should be writable by only the administrators on the system.\n",
    "references": "1. [https://kubernetes.io/docs/admin/kube-apiserver/](https://kubernetes.io/docs/admin/kube-apiserver/)\n",
    "remediation": "Run the below command (based on the file location on your system) on the\ncontrol plane node.\nFor example,\n```\nchmod 644 /etc/kubernetes/manifests/kube-apiserver.yaml\n```\n",
    "section": "Control Plane Node Configuration Files",
    "tags": [
      "CIS",
      "Kubernetes",
      "CIS 1.1.1",
      "Control Plane Node Configuration Files"
    ],
    "version": "1.0"
  }
}
````

</details>

### Evaluate with input schema

```bash
opa eval data.main --format pretty -i input.json -b ./bundle -s bundle/compliance/cis_k8s/schemas/input_schema.json
1 error occurred: bundle/compliance/lib/data_adapter.rego:11: rego_type_error: undefined ref: input.filenames
        input.filenames
              ^
              have: "filenames"
              want (one of): ["command" "filename" "gid" "mode" "path" "type" "uid"]

```

## Local Testing

### Test entire policy

```bash
opa build -b ./bundle -e ./bundle/compliance
```

```bash
opa test -b bundle.tar.gz -v
```

### Test specific rule

```bash
opa test -v bundle/compliance/kubernetes_common bundle/compliance/lib bundle/compliance/cis_k8s/test_data.rego bundle/compliance/cis_k8s/rules/cis_1_1_2 --ignore="common_tests.rego"
```

### Pre-commit hooks

see [pre-commit](https://pre-commit.com/) package

- Install the package `brew install pre-commit`
- Then run `pre-commit install`
- Finally `pre-commit run --all-files --verbose`

### Running opa server with the compliance policy

```bash
docker run --rm -p 8181:8181 -v $(pwd):/bundle openpolicyagent/opa:0.36.1 run -s -b /bundle
```

Test it 🚀

```bash
curl --location --request POST 'http://localhost:8181/v1/data/main' \
--header 'Content-Type: application/json' \
--data-raw '{
    "input": {
        "type": "file",
        "resource": {
            "type": "file",
            "mode": "700",
            "path": "/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml",
            "uid": "etc",
            "name": "kube-apiserver.yaml",
            "group": "root"
        }
    }
}'
```

### Adding new rules

Add a new rule package to `/bundle/compliance/<benchmark>/rules/<rule_name>`

1. Add `rule.rego` file that will contain the rule evaluation logic.
2. Add `test.rego` file that will contain the rule tests.
3. Generate rule metadata (`data.yaml`) and templates following the steps in the [README](dev/README.md)
