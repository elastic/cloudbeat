# Cloud Security Posture - Rego policies

[![CIS K8S](https://img.shields.io/badge/CIS-Kubernetes%20(74%25)-326CE5?logo=Kubernetes)](RULES.md#k8s-cis-benchmark)
[![CIS EKS](https://img.shields.io/badge/CIS-Amazon%20EKS%20(60%25)-FF9900?logo=Amazon+EKS)](RULES.md#eks-cis-benchmark)
[![CIS AWS](https://img.shields.io/badge/CIS-AWS%20(87%25)-232F3E?logo=Amazon+AWS)](RULES.md#aws-cis-benchmark)
[![CIS GCP](https://img.shields.io/badge/CIS-GCP%20(85%25)-4285F4?logo=Google+Cloud)](RULES.md#gcp-cis-benchmark)
[![CIS AZURE](https://img.shields.io/badge/CIS-AZURE%20(48%25)-0078D4?logo=Microsoft+Azure)](RULES.md#azure-cis-benchmark)

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
      "containers": [
        {
          "name": "aws-node",
          "securityContext": {
            "capabilities": {
              "add": ["NET_ADMIN"]
            }
          }
        }
      ]
    }
  },
  "rule": {
    "audit": "Get the set of PSPs with the following command:\n\n```\nkubectl get psp\n```\n\nFor each PSP, check whether capabilities have been forbidden:\n\n```\nkubectl get psp \u003cname\u003e -o=jsonpath='{.spec.requiredDropCapabilities}'\n```",
    "benchmark": {
      "id": "cis_eks",
      "name": "CIS Amazon Elastic Kubernetes Service (EKS)",
      "rule_number": "4.2.9",
      "version": "v1.0.1"
    },
    "default_value": "By default, PodSecurityPolicies are not defined.\n",
    "description": "Do not generally permit containers with capabilities",
    "id": "b28f5d7c-3db2-58cf-8704-b8e922e236b7",
    "impact": "Pods with containers require capabilities to operate will not be permitted.",
    "name": "Minimize the admission of containers with capabilities assigned",
    "profile_applicability": "* Level 2",
    "rationale": "Containers run with a default set of capabilities as assigned by the Container Runtime.\nCapabilities are parts of the rights generally granted on a Linux system to the root user.\n\nIn many cases applications running in containers do not require any capabilities to operate, so from the perspective of the principal of least privilege use of capabilities should be minimized.",
    "references": "1. https://kubernetes.io/docs/concepts/policy/pod-security-policy/#enabling-pod-security-policies\n2. https://www.nccgroup.trust/uk/our-research/abusing-privileged-and-unprivileged-linux-containers/",
    "remediation": "Review the use of capabilites in applications runnning on your cluster.\nWhere a namespace contains applicaions which do not require any Linux capabities to operate consider adding a PSP which forbids the admission of containers which do not drop all capabilities.",
    "section": "Pod Security Policies",
    "tags": ["CIS", "EKS", "CIS 4.2.9", "Pod Security Policies"],
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
opa test -v bundle --run 'cis_4_1.test'  # Test the 4.1 rule
opa test -v bundle --run 'cis_(4|5)'     # Test all rules of CIS section 4 and 5
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
