package compliance.policy.kube_api.ensure_external_ip

import data.compliance.lib.common as lib_common
import data.compliance.policy.kube_api.data_adapter
import future.keywords.if

verify_external_ip if {
	some address
	data_adapter.status.addresses[address].type == "ExternalIP"
	data_adapter.status.addresses[address].address != "0.0.0.0"
}

evidence["external_ip"] := result if {
	not data.rule_evaluation
	data_adapter.status.addresses[address].type == "ExternalIP"
	result = data_adapter.status.addresses[address]
}

finding(rule_evaluation) := result if {
	data_adapter.is_kube_node

	result_evidence = evidence with data.rule_evaluation as rule_evaluation

	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		result_evidence,
	)
}
