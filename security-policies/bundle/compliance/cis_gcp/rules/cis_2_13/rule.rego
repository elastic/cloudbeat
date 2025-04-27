package compliance.cis_gcp.rules.cis_2_13

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if
import future.keywords.in

# Ensure Cloud Asset Inventory Is Enabled
finding := result if {
	data_adapter.is_services_usage

	passed := [r | r := input.resource.services[_]; is_enabled_inventory_service(r)]
	failed := [r | r := input.resource.services[_]; not is_enabled_inventory_service(r)]
	ok := count(passed) > 0

	result := common.generate_result_without_expected(
		common.calculate_result(ok),
		{"services": {json.filter(c, ["resource/data"]) | c := common.get_evidence(ok, passed, failed)[_]}}
	)
}


is_enabled_inventory_service(service) if {
	service.resource.data.name == "cloudasset.googleapis.com"
	service.resource.data.state == "ENABLED"
} else := false
