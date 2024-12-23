package compliance.cis_aws.rules.cis_1_9

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if

default rule_evaluation := false

# Ensure that the number of previous passwords that IAM users are prevented from reusing is 24.
finding := result if {
	# filter
	data_adapter.is_pwd_policy

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(data_adapter.pwd_policy.reuse_prevention_count == 24),
		{"Password Policy:": data_adapter.pwd_policy},
	)
}
