package compliance.cis_k8s.rules.cis_5_2_7

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Minimize the admission of containers with the NET_RAW capability (Automated)

# evaluate
default rule_evaluation = false

# Verify that there is at least one PSP which returns NET_RAW or ALL.
rule_evaluation {
	# Verify that there is at least one PSP which returns NET_RAW.
	data_adapter.pod.spec.requiredDropCapabilities[_] == "NET_RAW"
}

# or 
rule_evaluation {
	# Verify that there is at least one PSP which returns ALL.
	data_adapter.pod.spec.requiredDropCapabilities[_] == "ALL"
}

finding = result {
	# filter
	data_adapter.is_kube_api

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": json.filter(data_adapter.pod, ["uid", "spec/requiredDropCapabilities"]),
	}
}

metadata = {
	"name": "Minimize the admission of containers with the NET_RAW capability",
	"description": "Do not generally permit containers with the potentially dangerous NET_RAW capability.",
	"rationale": `Containers run with a default set of capabilities as assigned by the Container Runtime.
By default this can include potentially dangerous capabilities.
With Docker as the container runtime the NET_RAW capability is enabled which may be misused by malicious containers.
Ideally, all containers should drop this capability.
There should be at least one PodSecurityPolicy (PSP) defined which prevents containers with the NET_RAW capability from launching.
If you need to run containers with this capability,
this should be defined in a separate PSP and you should carefully check RBAC controls to ensure that only limited service accounts and users are given permission to access that PSP.`,
	"impact": "Pods with containers which run with the NET_RAW capability will not be permitted.",
	"remediation": "Create a PSP as described in the Kubernetes documentation, ensuring that the .spec.requiredDropCapabilities is set to include either NET_RAW or ALL.",
	"default_value": "By default, PodSecurityPolicies are not defined.",
	"benchmark": cis_k8s.benchmark_metadata,
	"tags": array.concat(cis_k8s.default_tags, ["CIS 5.2.7", "Pod Security Policies"]),
}
