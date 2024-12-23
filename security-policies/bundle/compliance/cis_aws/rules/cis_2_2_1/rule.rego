package compliance.cis_aws.rules.cis_2_2_1

import data.compliance.lib.common
import data.compliance.policy.aws_ec2.data_adapter
import future.keywords.if

default rule_evaluation := false

# Ensure EBS Volume Encryption is Enabled in all Regions
finding := result if {
	# filter
	data_adapter.is_ebs_policy

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		input.resource,
	)
}

rule_evaluation if {
	input.resource.enabled == true
}
