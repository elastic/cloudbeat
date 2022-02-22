package compliance.cis_k8s.rules.cis_5_2_9

import data.compliance.cis_k8s
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
	"remediation": `Review the use of capabilites in applications runnning on your cluster.
Where a namespace contains applicaions which do not require any Linux capabities to operate consider adding a PSP which forbids the admission of containers which do not drop all capabilities.`,
	"default_value": "By default, PodSecurityPolicies are not defined.",
	"benchmark": cis_k8s.benchmark_metadata,
	"tags": array.concat(cis_k8s.default_tags, ["CIS 5.2.9", "Pod Security Policies"]),
}
