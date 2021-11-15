# Cloud Security Posture security policies 
    .
    ├── compliance                         # Compliance policies
    │   ├── lib
    │   │   ├── common.rego                # Common functions
    │   │   ├── data_adapter.rego          # Input data adapter
    │   │   └── test.rego                  # Common Test functions
    │   ├── rules/cis
    │   │   ├── cis_1_1_1                  # rule package 
    │   │   │   ├── rule.rego
    │   │   │   └── test.rego
    │   │   └── ...
    │   └── cis_k8s.rego                   # Handles all Kubernetes CIS rules evalutations
    └── main.rego                          # Evaluate all policies and returns the findings
    
## Local Evaluation
Add the following configuration files into the root folder
##### `data.yaml`
should contain the list of rules you want to evaluate (also supports json)

```yaml
activated_rules:
  cis_k8s:
    cis_1_1_1: true
    cis_1_1_2: true
```

##### `input.json`
should contain an beat/agent output, e.g. filesystem data

```json
{
    "type": "filesystem",
    "mode": "0700",
    "path": "/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml",
    "uid": "etc",
    "filename": "kube-apiserver.yaml",
    "gid": "root"
}
```

### Evaluate entire policy into output.json
`opa eval data.main --format pretty -i input.json -b . > output.json`

### Evaluate findings only
`opa eval data.main.findings --format pretty -i input.json -b . > output.json`

<details> 
<summary>Example output</summary>
  
```json
[
  {
    "evaluation": "violation",
    "evidence": {
      "filemode": "0700"
    },
    "rule_name": "Ensure that the API server pod specification file permissions are set to 644 or more restrictive",
    "tags": [
      "CIS",
      "CIS v1.6.0",
      "Kubernetes",
      "CIS 1.1.1"
    ]
  },
  {
    "evaluation": "violation",
    "evidence": {
      "gid": "root",
      "uid": "etc"
    },
    "rule_name": "Ensure that the API server pod specification file ownership is set to root:root",
    "tags": [
      "CIS",
      "CIS v1.6.0",
      "Kubernetes",
      "CIS 1.1.2"
    ]
  }
]


```
  
</details>

## Local Testing
### Test entire policy
`opa test -v compliance`

### Test specific rule
`opa test -v compliance/lib compliance/cis_k8s.rego compliance/rules/cis_1_1_2`

### Pre-commit hooks
see [pre-commit](https://pre-commit.com/) package

- Install the package `brew install pre-commit`
- Then run `pre-commit install`
- Finally `pre-commit run --all-files --verbose`
