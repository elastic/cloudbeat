package compliance.policy.gcp.iam.ensure_policy_not_managed_by_user

import data.compliance.policy.gcp.data_adapter
import future.keywords.every
import future.keywords.if

default is_policy_not_managed_by_user := false

is_policy_not_managed_by_user if {
	every member in data_adapter.iam_policy.bindings[i].members {
		not startswith(member, "user:")
	}
}
