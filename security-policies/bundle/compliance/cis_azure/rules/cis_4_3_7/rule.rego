package compliance.cis_azure.rules.cis_4_3_7

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_postgresql_server_db

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(firewall_rules_properly_configured),
		{"Resource": data_adapter.resource},
	)
}

default firewall_rules_properly_configured := false

firewall_rules_properly_configured if {
	not has_allow_all_firewall_rule
}

has_allow_all_firewall_rule if {
	some i
	data_adapter.resource.extension.psqlFirewallRules[i].name == "AllowAllWindowsAzureIps"
}

has_allow_all_firewall_rule if {
	some i
	data_adapter.resource.extension.psqlFirewallRules[i].properties.startIPAddress == "0.0.0.0"
	data_adapter.resource.extension.psqlFirewallRules[i].properties.endIPAddress == "0.0.0.0"
}
