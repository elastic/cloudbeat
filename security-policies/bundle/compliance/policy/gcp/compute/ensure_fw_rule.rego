package compliance.policy.gcp.compute.ensure_fw_rule

import data.compliance.lib.assert
import data.compliance.policy.gcp.data_adapter
import future.keywords.every
import future.keywords.if
import future.keywords.in

is_valid_fw_rule(port) := false if {
	some range in data_adapter.resource.data.sourceRanges
	range == "0.0.0.0/0"
	data_adapter.resource.data.direction == "INGRESS"

	some action in data_adapter.resource.data.allowed
	action.IPProtocol in {"tcp", "all"}
	is_port_effective(port, object.get(action, ["ports"], []))
} else := true

# The ports list can include both ranges, such as 80-90, and individual ports, such as 443.
is_port_effective(port, ports) if {
	# If the ports list is empty, then the rule is effective for all ports.
	assert.array_is_empty(ports)
} else if {
	some port_exp in ports
	to_number(port_exp) == port
} else if {
	is_port_within_range(port, ports)
}

# Check if a port is within the specified ranges
is_port_within_range(port, ports) if {
	some port_exp in ports
	parts = split(port_exp, "-")
	start_port = to_number(parts[0])
	end_port = to_number(parts[1])

	start_port <= port
	end_port >= port
}
