package compliance.cis_gcp.rules.cis_1_8

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import rego.v1

# a user should not have both admin and user role
# this creates a set of such members, and
# if the set is empty, the policy is valid
members_with_both_roles contains m if {
	# get all members with admin role
	some admin in data_adapter.iam_policy.bindings
	admin.role == "roles/iam.serviceAccountAdmin"
	some m in admin.members

	# get all members with user role
	some user in data_adapter.iam_policy.bindings
	user.role == "roles/iam.serviceAccountUser"
	m in user.members
}

finding = result if {
	data_adapter.is_cloud_resource_manager_project
	data_adapter.has_policy

	no_admin_with_user_role := count(members_with_both_roles) == 0

	result := common.generate_result_without_expected(
		common.calculate_result(no_admin_with_user_role),
		data_adapter.iam_policy,
	)
}
