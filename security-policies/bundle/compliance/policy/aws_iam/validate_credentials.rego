package compliance.policy.aws_iam.validate_credentials

import data.compliance.lib.assert
import data.compliance.policy.aws_iam.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if

duration := sprintf("%dh", [45 * 24]) # 45 days converted to hours

default validate_credentials := false

# checks if the user has a password enabled and if the user's last access date is within the specified duration.
validate_credentials if {
	assert.array_is_empty(data_adapter.active_access_keys)
	data_adapter.iam_user.password_enabled
	common.are_credentials_within_duration([data_adapter.iam_user], "last_access", duration)
}

# checks if the user does not have a password enabled and if all of the user's active access keys are valid.
validate_credentials if {
	not data_adapter.iam_user.password_enabled
	validate_access_keys
}

# checks if the user has a password enabled, if the user's last access date is within the specified duration, and if all of the user's active access keys are valid
validate_credentials if {
	data_adapter.iam_user.password_enabled
	common.are_credentials_within_duration([data_adapter.iam_user], "last_access", duration)
	validate_access_keys
}

# checks if the user has a password enabled, if the user's last access date is "no_information", the user's password last changed date should be within the specified duration.
# In addition checks that all of the user's active access keys are valid
validate_credentials if {
	data_adapter.iam_user.password_enabled
	data_adapter.iam_user.last_access == "no_information"
	common.are_credentials_within_duration([data_adapter.iam_user], "password_last_changed", duration)
	validate_access_keys
}

# The first rule checks if the last access date of the used active access keys is within the specified duration.
# The second rule checks if the rotation date of the unused active access keys is within the specified duration.
validate_access_keys if {
	common.are_credentials_within_duration(data_adapter.used_active_access_keys, "last_access", duration)
	common.are_credentials_within_duration(data_adapter.unused_active_access_keys, "rotation_date", duration)
}
