package compliance.policy.kube_api.minimize_wildcard

import data.compliance.lib.assert
import data.compliance.lib.common as lib_common
import data.compliance.policy.kube_api.data_adapter
import future.keywords.if
import future.keywords.in

default rule_violation := false

rule_violation if {
	cluster_roles_rule := data_adapter.cluster_roles.rules[i]
	is_using_wildcards(cluster_roles_rule)
}

finding := result if {
	data_adapter.is_cluster_roles

	rule_evaluation := assert.is_false(rule_violation)

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"cluster_roles": data_adapter.cluster_roles},
	)
}

is_using_wildcards(rule) if {
	"*" in rule.apiGroups # assert no wild-cards in api_group
}

is_using_wildcards(rule) if {
	"*" in rule.resources # assert no wild-cards in resources
}

is_using_wildcards(rule) if {
	"*" in rule.verbs # assert no wild-cards in verbs
}
