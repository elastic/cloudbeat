package compliance.cis_eks.rules.cis_4_2_9

import data.compliance.cis_eks
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Minimize the admission of containers with capabilities assigned (Manual)
# evaluate
default rule_evaluation = true

rule_evaluation = false {
	container := data_adapter.containers[_]
	capabilities := object.get(container.securityContext, "capabilities", [])
	not assert.array_is_empty(capabilities)
}

finding = result {
	# filter
	data_adapter.is_kube_api
	data_adapter.containers

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"containers": {json.filter(c, ["name", "securityContext/capabilities"]) | c := data_adapter.containers[_]},
	}
}

metadata = {
	"name": "Minimize the admission of containers with capabilities assigned",
	"description": "Do not generally permit containers with capabilities",
	"rationale": `Containers run with a default set of capabilities as assigned by the Container Runtime.
Capabilities are parts of the rights generally granted on a Linux system to the root user.
In many cases applications running in containers do not require any capabilities to operate,
so from the perspective of the principal of least privilege use of capabilities should be minimized.`,
	"impact": "Pods with containers require capabilities to operate will not be permitted.",
	"remediation": `Review the use of capabilities in applications running on your cluster.
Where a namespace contains applications which do not require any Linux capabilities to operate consider adding a PSP which forbids the admission of containers which do not drop all capabilities.`,
	"default_value": "By default, PodSecurityPolicies are not defined.",
	"benchmark": cis_eks.benchmark_metadata,
	"tags": array.concat(cis_eks.default_tags, ["CIS 4.2.9", "Pod Security Policies"]),
}
