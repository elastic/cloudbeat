package compliance.cis_azure.rules.cis_1_23

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if
import future.keywords.in

finding := result if {
	# filter
	data_adapter.is_custom_role_definition

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(evaluation_results),
		{"Resource": data_adapter.resource},
	)
}

has_administrator_action if {
	some permission in data_adapter.properties.permissions
	some action in permission.actions
	action == "*"
}

has_subscription_scope if {
	some scope in data_adapter.properties.assignableScopes
	regex.match(`^/(?:subscriptions/[a-z\d]{8}-[a-z\d]{4}-[a-z\d]{4}-[a-z\d]{4}-[a-z\d]{12})?$`, scope)
}

has_administrator_subscription_scope if {
	has_subscription_scope
	has_administrator_action
}

evaluation_results if {
	not has_administrator_subscription_scope
} else := false
