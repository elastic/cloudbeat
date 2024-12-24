package compliance.policy.gcp.iam.ensure_no_public_access

import data.compliance.policy.gcp.data_adapter
import future.keywords.if

default resource_is_public := false

resource_is_public if {
	# Check if the IAM policy is not empty
	data_adapter.iam_policy
	some i, j
	data_adapter.iam_policy.bindings[i].members[j] == "allUsers"
}

resource_is_public if {
	# Check if the IAM policy is not empty
	data_adapter.iam_policy
	some i, j
	data_adapter.iam_policy.bindings[i].members[j] == "allAuthenticatedUsers"
}
