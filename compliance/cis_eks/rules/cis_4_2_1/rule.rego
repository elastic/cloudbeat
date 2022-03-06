package compliance.cis_eks.rules.cis_4_2_1

import data.compliance.cis_eks
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Minimize the admission of privileged containers (Automated)

# evaluate
default rule_evaluation = true

# Verify that there is at least one PSP which does not return true.
rule_evaluation = false {
	container := data_adapter.containers[_]
	common.contains_key_with_value(container.securityContext, "privileged", true)
}

finding = result {
	# filter
	data_adapter.is_kube_api

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {
			"uid": data_adapter.pod.uid,
			"containers": {json.filter(c, ["name", "securityContext/privileged"]) | c := data_adapter.containers[_]},
		},
	}
}

metadata = {
	"name": "Minimize the admission of privileged containers",
	"description": "Do not generally permit containers to be run with the securityContext.privileged flag set to true.",
	"rationale": `Privileged containers have access to all Linux Kernel capabilities and devices.
A container running with full privileges can do almost everything that the host can do.
This flag exists to allow special use-cases, like manipulating the network stack and accessing devices.
There should be at least one PodSecurityPolicy (PSP) defined which does not permit privileged containers.
If you need to run privileged containers, this should be defined in a separate PSP and you should carefully check RBAC controls to ensure that only limited service accounts and users are given permission to access that PSP.`,
	"impact": "Pods defined with spec.containers[].securityContext.privileged: true will not be permitted.",
	"remediation": "Create a PSP as described in the Kubernetes documentation, ensuring that the .spec.privileged field is omitted or set to false.",
	"default_value": "By default, when you provision an EKS cluster, a pod security policy called eks.privileged is automatically created.",
	"benchmark": cis_eks.benchmark_metadata,
	"tags": array.concat(cis_eks.default_tags, ["CIS 4.2.1", "Pod Security Policies"]),
}
