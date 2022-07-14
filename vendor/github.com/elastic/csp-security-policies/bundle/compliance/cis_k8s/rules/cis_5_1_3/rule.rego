package compliance.cis_k8s.rules.cis_5_1_3

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter
import future.keywords.in

# Minimize wildcard use in Roles and ClusterRoles (Manual)
default rule_violation = false

# evaluate
rule_violation {
	cluster_roles_rule := data_adapter.cluster_roles.rules[i]
	is_using_wildcards(cluster_roles_rule)
}

finding = result {
	# filter
	data_adapter.is_cluster_roles

	# evaluate
	rule_evaluation := assert.is_false(rule_violation)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"cluster_roles": data_adapter.cluster_roles},
	}
}

is_using_wildcards(rule) {
	"*" in rule.apiGroups # assert no wild-cards in api_group
}

is_using_wildcards(rule) {
	"*" in rule.resources # assert no wild-cards in resources
}

is_using_wildcards(rule) {
	"*" in rule.verbs # assert no wild-cards in verbs
}
