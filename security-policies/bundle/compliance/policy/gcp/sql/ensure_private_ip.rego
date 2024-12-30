package compliance.policy.gcp.sql.ensure_private_ip

import data.compliance.policy.gcp.data_adapter
import future.keywords.every
import future.keywords.if

ip_is_private if {
	every ipAddress in data_adapter.resource.data.ipAddresses {
		not ipAddress.type == "PRIMARY"
	}
} else := false
