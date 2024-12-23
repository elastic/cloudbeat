package compliance.policy.aws_iam.ensure_hardware_mfa

import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if

default ensure_hardware_mfa_device := false

# Only one MFA device can be received as input,
# even if a user has multiple MFA devices linked to their account.
# This limitation is because we are unable to collect information about hardware MFA devices.
# If a software MFA device is available, it will be the one that is received as input.
# If no software MFA device is available and MFA is enabled, an MFA hardware device will be received instead.
ensure_hardware_mfa_device if {
	data_adapter.iam_user.mfa_active
	device := data_adapter.iam_user.mfa_devices[_]
	not device.is_virtual
}
