package compliance.cis_k8s.rules.cis_5_1_6

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
	data_adapter.is_kube_api
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
