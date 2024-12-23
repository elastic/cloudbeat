package compliance.policy.aws_iam.ensure_access_keys_use

import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if

default ensure_access_keys_use := true

ensure_access_keys_use := false if {
	data_adapter.iam_user.password_enabled
	key := data_adapter.active_access_keys[_]
	not key.has_used
}
