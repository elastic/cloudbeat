package compliance.policy.aws_iam.verify_keys_rotation

import data.compliance.policy.aws_iam.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if

duration := sprintf("%dh", [90 * 24]) # 90 days converted to hours

default verify_rotation := false

verify_rotation if {
	common.are_credentials_within_duration(data_adapter.active_access_keys, "rotation_date", duration)
}
