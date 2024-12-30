package compliance.policy.aws_iam.verify_user_usage

import data.compliance.policy.aws_iam.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if

default verify_user_usage := false

verify_user_usage if {
	not common.are_credentials_within_duration(data_adapter.active_access_keys, "last_access", "24h")
	not common.are_credentials_within_duration([data_adapter.iam_user], "last_access", "24h")
}

verify_user_usage if {
	count(data_adapter.active_access_keys) == 0
	not common.are_credentials_within_duration([data_adapter.iam_user], "last_access", "24h")
}
