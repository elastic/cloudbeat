package compliance.cis_k8s.rules.cis_5_1_6

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter
import future.keywords.in

# Ensure that default service accounts are not actively used (Manual)
default rule_violation = false

# Review pod and service account objects in the cluster and ensure that automountServiceAccountToken is
# set to false
rule_violation {
	pod := data_adapter.pod
	pod.spec.automountServiceAccountToken == true
}

rule_violation {
	service_account := data_adapter.service_account
	service_account.automountServiceAccountToken == true
}

finding = result {
	# filter
	data_adapter.is_service_account_or_pod

	# evaluate
	rule_evaluation = assert.is_false(rule_violation)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {
			"serviceAccounts": [pod.spec.serviceAccount | data_adapter.pods[pod]],
			"serviceAccount": [service_account | data_adapter.service_accounts[service_account]],
		},
	}
}

metadata = {
	"name": "Ensure that default service accounts are not actively used",
	"description": `Service accounts tokens should not be mounted in pods except where the workload
running in the pod explicitly needs to communicate with the API server.`,
	"rationale": `Mounting service account tokens inside pods can provide an avenue for privilege escalation
attacks where an attacker is able to compromise a single pod in the cluster.
Avoiding mounting these tokens removes this attack avenue.`,
	"impact": `Pods mounted without service account tokens will not be able to communicate with the API
server, except where the resource is available to unauthenticated principals.`,
	"remediation": `Modify the definition of pods and service accounts which do not need to mount service
account tokens to disable it.`,
	"default_value": "By default, all pods get a service account token mounted in them.",
	"benchmark": cis_k8s.benchmark_metadata,
	"tags": array.concat(cis_k8s.default_tags, ["CIS 5.1.6", "RBAC and Service Accounts"]),
}
