package compliance.policy.gcp.compute.assess_instance_metadata

import data.compliance.policy.gcp.data_adapter
import future.keywords.every
import future.keywords.if

is_instance_metadata_valid(key, expected_val) if {
	some item in data_adapter.resource.data.metadata.items
	item.key == key
	item.value == expected_val
} else := false
