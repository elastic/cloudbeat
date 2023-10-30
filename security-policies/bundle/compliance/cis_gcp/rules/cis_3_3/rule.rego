package compliance.cis_gcp.rules.cis_3_3

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter

default is_dnssec_enabled = false

# Ensure That DNSSEC Is Enabled for Cloud DNS.
finding = result {
	# filter
	data_adapter.is_dns_managed_zone

	# only apply to public zones
	data_adapter.resource.data.visibility == "PUBLIC"

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_dnssec_enabled),
		{"Managed zone": input.resource},
	)
}

is_dnssec_enabled {
	data_adapter.resource.data.dnssecConfig.state == "ON"
}
