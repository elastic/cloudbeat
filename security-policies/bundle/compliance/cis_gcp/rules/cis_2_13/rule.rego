package compliance.cis_gcp.rules.cis_2_13

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if
import future.keywords.in

# Ensure Cloud Asset Inventory Is Enabled
finding := result if {
	data_adapter.is_services_usage

	result := common.generate_result_without_expected(
		common.calculate_result(is_asset_inventory_enabled),
		input.resource.services,
	)
}

is_asset_inventory_enabled if {
	some service in input.resource.services
	service.resource.data.name == "cloudasset.googleapis.com"
	service.resource.data.state == "ENABLED"
} else := false
