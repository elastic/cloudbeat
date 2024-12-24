package compliance.policy.aws_rds.ensure_no_public_access

import data.compliance.lib.common as lib_common
import data.compliance.policy.aws_rds.data_adapter
import future.keywords.if
import future.keywords.in

default has_public_access := false

has_public_access if {
	data_adapter.publicly_accessible == true
	subnets := data_adapter.subnets[_]
	some route in subnets.RouteTable.Routes
	route.DestinationCidrBlock == "0.0.0.0/0"
	startswith(route.GatewayId, "igw-")
}

has_subnets_without_route_table if {
	subnets := data_adapter.subnets[_]
	subnets != []
	subnets.RouteTable == null
}

finding := result if {
	data_adapter.is_rds
	not has_subnets_without_route_table

	rule_evaluation := has_public_access == false
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"PubliclyAccessible": data_adapter.publicly_accessible, "Subnets": data_adapter.subnets},
	)
}
