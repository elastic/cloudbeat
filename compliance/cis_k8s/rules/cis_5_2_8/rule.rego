package compliance.cis_k8s.rules.cis_5_2_8

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Minimize the admission of containers with added capabilities (Automated)

finding = result {
	# filter
	data_adapter.is_kube_api

	# evaluate
	allowedCapabilities := object.get(data_adapter.pod.spec, "allowedCapabilities", [])
	rule_evaluation := assert.array_is_empty(allowedCapabilities)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": json.filter(data_adapter.pod, ["uid", "spec/allowedCapabilities"]),
	}
}

metadata = {
	"name": "Minimize the admission of containers with added capabilities",
	"description": "Do not generally permit containers with capabilities assigned beyond the default set.",
	"rationale": `Containers run with a default set of capabilities as assigned by the Container Runtime.
Capabilities outside this set can be added to containers which could expose them to risks of container breakout attacks.
There should be at least one PodSecurityPolicy (PSP) defined which prevents containers with capabilities beyond the default set from launching.
If you need to run containers with additional capabilities, this should be defined in a separate PSP and you should carefully check RBAC controls to ensure that only limited service accounts and users are given permission to access that PSP.`,
	"impact": "Pods with containers which require capabilities outwith the default set will not be permitted.",
	"remediation": "Ensure that allowedCapabilities is not present in PSPs for the cluster unless it is set to an empty array.",
	"default_value": "By default, PodSecurityPolicies are not defined.",
	"benchmark": cis_k8s.benchmark_name,
	"tags": array.concat(cis_k8s.default_tags, ["CIS 5.2.8", "Pod Security Policies"]),
}
