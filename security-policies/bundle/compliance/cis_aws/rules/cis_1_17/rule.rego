package compliance.cis_aws.rules.cis_1_17

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if
import future.keywords.in

# Ensure a support role has been created to manage incidents with AWS Support
finding := result if {
	# filter
	data_adapter.is_aws_support_access

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(aws_support_has_attached_roles),
		input.resource,
	)
}

aws_support_has_attached_roles if {
	# Implicitly check that the "roles" array exists and it's not empty. Check that the RoleId is not an empty string as
	# a sanity test.
	some role in data_adapter.roles
	role.RoleId != ""
} else := false
