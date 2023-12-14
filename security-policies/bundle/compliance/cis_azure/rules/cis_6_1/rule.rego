package compliance.cis_azure.rules.cis_6_1

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

targetPort := "3389"

targetPortInt := to_number(targetPort)

finding = result if {
	# filter
	data_adapter.is_vm

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(has_no_misconfiguration),
		{"Resource": data_adapter.resource},
	)
}

default has_no_misconfiguration = false

has_no_misconfiguration if {
	not has_rdp_misconfiguration
}

default has_rdp_misconfiguration = false

has_rdp_misconfiguration if {
	some i
	rule := data_adapter.resource.extension.network.securityRules[i]
	rule.access == "Allow"
	rule.protocol == "TCP"
	rule.direction == "Inbound"

	is_destination_range_rdp
	is_source_too_open
}

default is_destination_range_rdp = false

# Definition of port range:
#   - A single port, such as 80;
#   - A port range, such as 1024-65535;
#   - A comma-separated list of single ports and/or port ranges, such as 80,1024-65535.
#   - An asterisk (*) to allow traffic on any port.

is_rdp_port("*") := true

is_rdp_port(p) if {
	# split single ports and ranges
	portRanges := split(p, ",")

	some i
	is_in_range(portRanges[i])
}

# Check if it's a single port or range explicitly mentions targetPort
is_in_range(portRange) if {
	contains(portRange, targetPort)
}

# Check if the range contains port targetPort
is_in_range(portRange) if {
	contains(portRange, "-")
	boundaries := split(portRange, "-")
	to_number(boundaries[0]) <= targetPortInt
	to_number(boundaries[1]) >= targetPortInt
}

is_destination_range_rdp if {
	some i
	is_rdp_port(data_adapter.resource.extension.network.securityRules[i].destinationPortRange)
}

is_destination_range_rdp if {
	some i, j
	is_rdp_port(data_adapter.resource.extension.network.securityRules[i].destinationPortRanges[j])
}

default is_source_too_open = false

is_bad_address("*") := true

is_bad_address("0.0.0.0") := true

is_bad_address("<nw>/0") := true

is_bad_address("internet") := true

is_bad_address("any") := true

is_source_too_open if {
	some i
	is_bad_address(data_adapter.resource.extension.network.securityRules[i].sourceAddressPrefix)
}

is_source_too_open if {
	some i, j
	is_bad_address(data_adapter.resource.extension.network.securityRules[i].sourceAddressPrefixes[j])
}
