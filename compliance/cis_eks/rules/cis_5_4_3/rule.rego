package compliance.cis_eks.rules.cis_5_4_3

import data.compliance.lib.common
import data.compliance.lib.data_adapter

default rule_evaluation = true

# Verify that the node doesn't have an external IP
rule_evaluation = false {
	some address
	input.resource.status.addresses[address].type == "ExternalIP"
	input.resource.status.addresses[address].address != "0.0.0.0"
}

evidence["external_ip"] = result {
	not rule_evaluation
	input.resource.status.addresses[address].type == "ExternalIP"
	result = input.resource.status.addresses[address]
}

# Ensure there cluster node don't have a public IP
finding = result {
	# filter
	data_adapter.is_kube_node

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": evidence,
	}
}
