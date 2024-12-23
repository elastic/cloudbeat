package compliance.cis_aws.rules.cis_1_8

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if

default rule_evaluation := false

finding := result if {
	# filter
	data_adapter.is_pwd_policy

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		{"Password Policy:": data_adapter.pwd_policy},
	)
}

rule_evaluation if {
	# verify password length is equal or above 14
	common.greater_or_equal(data_adapter.pwd_policy.minimum_length, 14)
}
