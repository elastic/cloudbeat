package compliance.cis_gcp.rules.cis_6_5

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if
import future.keywords.in

finding := result if {
	data_adapter.is_sql_instance

	result := common.generate_result_without_expected(
		common.calculate_result(assert.is_false(is_publicly_accessible)),
		data_adapter.resource,
	)
}

is_publicly_accessible if {
	networks := object.get(
		data_adapter.resource.data.settings,
		["ipConfiguration", "authorizedNetworks"],
		[{"value": ""}],
	)
	networks[i].value == "0.0.0.0/0"
} else := false
