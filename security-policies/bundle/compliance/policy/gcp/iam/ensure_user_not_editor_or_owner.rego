package compliance.policy.gcp.iam.ensure_user_not_editor_or_owner

import data.compliance.policy.gcp.data_adapter
import future.keywords.if

default is_user_owner_or_editor := false

is_user_owner_or_editor if {
	# at least one member that starts with "user:"
	member := data_adapter.iam_policy.bindings[i].members[_]
	startswith(member, "user:")

	# Ensure the role is not a service account managed by Google (iam.gserviceaccount.com suffix).
	role := data_adapter.iam_policy.bindings[i].role
	not endswith(role, "iam.gserviceaccount.com")

	# Confirm the role is not editor/owner
	role != "roles/editor"
	role != "roles/owner"
}
