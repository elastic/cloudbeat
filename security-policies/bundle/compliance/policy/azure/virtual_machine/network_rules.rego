package compliance.policy.azure.virtual_machine.network_rules

import future.keywords.if

vm_has_closed_port(vm, targetPort, protocol) if {
	not vm_has_open_port(vm, targetPort, protocol)
} else := false

vm_has_open_port(vm, targetPort, protocol) if {
	some i
	rule := vm.resource.extension.network.securityRules[i]
	rule.access == "Allow"
	rule.protocol == protocol
	rule.direction == "Inbound"

	network_has_port(rule, targetPort)
	is_source_address_too_open(rule)
}

network_has_port(rule, targetPort) if {
	is_port_included(rule.destinationPortRange, targetPort)
}

network_has_port(rule, targetPort) if {
	some i
	is_port_included(rule.destinationPortRanges[i], targetPort)
}

is_source_address_too_open(rule) if {
	is_any_address(rule.sourceAddressPrefix)
}

is_source_address_too_open(rule) if {
	some i
	is_any_address(rule.sourceAddressPrefixes[i])
}

is_port_included("*", _) := true

is_port_included(p, targetPort) if {
	# split single ports and ranges
	portRanges := split(p, ",")

	some i
	is_in_range(portRanges[i], targetPort)
}

# Definition of port range:
#   - A single port, such as 80;
#   - A port range, such as 1024-65535;
#   - A comma-separated list of single ports and/or port ranges, such as 80,1024-65535.
#   - An asterisk (*) to allow traffic on any port.

# Check if it's a single port or range explicitly mentions targetPort
is_in_range(portRange, targetPort) if {
	not contains(portRange, "-")
	portRange == targetPort
}

# Check if the range contains port targetPort
is_in_range(portRange, targetPort) if {
	contains(portRange, "-")
	boundaries := split(portRange, "-")
	to_number(boundaries[0]) <= to_number(targetPort)
	to_number(boundaries[1]) >= to_number(targetPort)
}

is_any_address("*") := true

is_any_address("0.0.0.0") := true

is_any_address("<nw>/0") := true

is_any_address("internet") := true

is_any_address("any") := true
