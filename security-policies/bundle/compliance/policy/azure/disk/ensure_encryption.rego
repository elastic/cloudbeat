package compliance.policy.azure.disk.ensure_encryption

import data.compliance.policy.azure.data_adapter

encryption_type = data_adapter.properties.encryption.type

default is_encryption_enabled = false

is_encryption_enabled {
	encryption_type == "EncryptionAtRestWithCustomerKey"
}

is_encryption_enabled {
	encryption_type == "EncryptionAtRestWithPlatformAndCustomerKeys"
}
