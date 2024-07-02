package compliance.policy.gcp.iam.ensure_admin_without_multiple_roles

import data.compliance.policy.gcp.data_adapter
import future.keywords.if
import future.keywords.in

admin_has_multiple_roles(admin_role, other_role) if {
	admin := data_adapter.iam_policy.bindings[_]
	admin.role == admin_role

	other := data_adapter.iam_policy.bindings[_]
	other.role == other_role

	m := admin.members[_]
	m in other.members
}
