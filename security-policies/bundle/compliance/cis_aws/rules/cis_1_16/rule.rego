package compliance.cis_aws.rules.cis_1_16

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if
import future.keywords.in

# Ensure IAM policies that allow full "*:*" administrative privileges are not attached
finding := result if {
	# filter
	data_adapter.is_iam_policy

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(policy_is_permissive == false),
		{"Statements": data_adapter.policy_document.Statement},
	)
}

policy_is_permissive if {
	some statement in data_adapter.policy_document.Statement
	statement.Effect == "Allow"
	"*" in common.ensure_array(statement.Action)
	"*" in common.ensure_array(statement.Resource)
} else := false
