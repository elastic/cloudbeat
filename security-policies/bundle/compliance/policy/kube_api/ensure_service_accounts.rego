package compliance.policy.kube_api.ensure_service_accounts

import data.compliance.lib.assert
import data.compliance.lib.common as lib_common
import data.compliance.policy.kube_api.data_adapter
import future.keywords.if

finding(rule_violation) := result if {
	data_adapter.is_kube_api
	data_adapter.is_service_account_or_pod

	rule_evaluation = assert.is_false(rule_violation)

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{
			"serviceAccounts": [pod.spec.serviceAccount | data_adapter.pods[pod]],
			"serviceAccount": [service_account | data_adapter.service_accounts[service_account]],
		},
	)
}

default service_account_automount := false

# Review pod and service account objects in the cluster and ensure that automountServiceAccountToken is
# set to false
service_account_automount if {
	pod := data_adapter.pod
	pod.spec.automountServiceAccountToken == true
}

service_account_automount if {
	service_account := data_adapter.service_account
	service_account.automountServiceAccountToken == true
}

default service_account_default := false

# no roles or cluster roles bound to default service account apart from the defaults.
service_account_default if {
	pod := data_adapter.pod
	pod.spec.serviceAccount == "default"
}

# no roles or cluster roles bound to default service account apart from the defaults.
service_account_default if {
	pod := data_adapter.pod
	pod.spec.serviceAccountName == "default"
}

# ensure that the automountServiceAccountToken: false
service_account_default if {
	service_account := data_adapter.service_account
	service_account.metadata.name == "default"
	service_account.automountServiceAccountToken == true
}
