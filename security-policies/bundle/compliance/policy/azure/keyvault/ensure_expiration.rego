package compliance.policy.azure.disk.ensure_expiration

import future.keywords.every
import future.keywords.if

all_enabled_items_have_expiration(items) if {
	enabled = [item | item := items[_]; item.properties.attributes.enabled == true]

	every item in enabled {
		item.properties.attributes.exp > 0
	}
} else := false
