package compliance.cis_azure.rules.cis_9_5

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

default rule_evaluation := false

finding := result if {
	# filter
	data_adapter.is_website_asset

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		data_adapter.resource,
	)
}

rule_evaluation if {
	data_adapter.identity.principalId != null
}
