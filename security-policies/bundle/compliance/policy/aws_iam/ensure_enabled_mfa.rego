package compliance.policy.aws_iam.ensure_enabled_mfa

import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if

default ensure_mfa_device := false

ensure_mfa_device if {
	data_adapter.iam_user.password_enabled
	data_adapter.iam_user.mfa_active
}

ensure_mfa_device if {
	not data_adapter.iam_user.password_enabled
	not data_adapter.iam_user.mfa_active
}
