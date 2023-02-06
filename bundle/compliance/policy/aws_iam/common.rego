package compliance.policy.aws_iam.common

import data.compliance.lib.common
import future.keywords.every

are_credentials_within_duration(keys, field, duration) {
	every key in keys {
		common.date_within_duration(time.parse_rfc3339_ns(key[field]), duration)
	}
} else = false
