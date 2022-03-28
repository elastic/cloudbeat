package compliance.cis_k8s.rules.cis_5_1_5

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter
import future.keywords.in

# Ensure that default service accounts are not actively used (Manual)
default rule_violation = false

# no roles or cluster roles bound to default service account apart from the defaults.
rule_violation {
	pod := data_adapter.pod
	pod.spec.serviceAccount == "default"
}

# no roles or cluster roles bound to default service account apart from the defaults.
rule_violation {
	pod := data_adapter.pod
	pod.spec.serviceAccountName == "default"
}

# ensure that the automountServiceAccountToken: false
rule_violation {
	service_account := data_adapter.service_account
	service_account.metadata.name == "default"
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
	"description": `The default service account should not be used to ensure that rights granted to
applications can be more easily audited and reviewed.`,
	"rationale": `Kubernetes provides a default service account which is used by cluster workloads where no specific service account is assigned to the pod.
Where access to the Kubernetes API from a pod is required, a specific service account should be created for that pod, and rights granted to that service account.
The default service account should be configured such that it does not provide a service account token and does not have any explicit rights assignments.`,
	"impact": `All workloads which require access to the Kubernetes API will require an explicit service
account to be created.`,
	"remediation": `Create explicit service accounts wherever a Kubernetes workload requires specific access to the Kubernetes API server.
Modify the configuration of each default service account to include this value automountServiceAccountToken: false`,
	"default_value": "By default the default service account allows for its service account token to be mounted in pods in its namespace.",
	"benchmark": cis_k8s.benchmark_metadata,
	"tags": array.concat(cis_k8s.default_tags, ["CIS 5.1.5", "RBAC and Service Accounts"]),
}
