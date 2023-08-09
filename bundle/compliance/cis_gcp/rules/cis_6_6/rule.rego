package compliance.cis_gcp.rules.cis_6_6

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

finding = result if {
	data_adapter.is_sql_instance
	is_clous_sql_instance_second_gen

	result := common.generate_result_without_expected(
		common.calculate_result(ip_is_private),
		data_adapter.resource,
	)
}

ip_is_private if {
	ips := data_adapter.resource.data.ipAddresses[i]
	ips.type != "PRIMARY"
	ips.type == "PRIVATE"
} else = false

is_clous_sql_instance_second_gen if {
	data_adapter.resource.data.instanceType == "CLOUD_SQL_INSTANCE"
	data_adapter.resource.data.backendType == "SECOND_GEN"
}
