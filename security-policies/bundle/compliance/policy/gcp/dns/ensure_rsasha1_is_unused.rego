package compliance.policy.gcp.dns.ensure_no_sha1

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if
import future.keywords.in

finding(type) := result if {
	# filter
	data_adapter.is_dns_managed_zone
	data_adapter.resource.data.visibility == "PUBLIC"

	# set result
	result := common.generate_evaluation_result(common.calculate_result(assert.is_false(is_sha1_used(type))))
}

is_sha1_used(type) if {
	some key_spec in data_adapter.resource.data.dnssecConfig.defaultKeySpecs
	key_spec.keyType == type
	key_spec.algorithm == "RSASHA1"
} else := false
