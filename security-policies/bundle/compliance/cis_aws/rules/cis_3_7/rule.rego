package compliance.cis_aws.rules.cis_3_7

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import future.keywords.if

default rule_evaluation := false

# Ensure CloudTrail logs are encrypted at rest using KMS CMKs.
finding := result if {
	# filter
	data_adapter.is_single_trail

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		input.resource,
	)
}

rule_evaluation if {
	not data_adapter.trail.KmsKeyId == null
	not data_adapter.trail.KmsKeyId == ""
}
