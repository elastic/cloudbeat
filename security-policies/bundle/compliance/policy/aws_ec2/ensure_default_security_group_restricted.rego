package compliance.policy.aws_ec2.ensure_default_security_group_restricted

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.policy.aws_ec2.data_adapter
import future.keywords.if

default rule_evaluation := false

finding := result if {
	# filter
	data_adapter.is_security_group_policy
	data_adapter.is_default_security_group

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		input.resource,
	)
}

# check if a security group is the default security group and if it has no inbound or outbound rules.
# non default security group can have any rules
rule_evaluation if {
	assert.all_true([
		count(data_adapter.security_group_inbound_rules) == 0,
		count(data_adapter.security_group_outbound_rules) == 0,
	])
}
