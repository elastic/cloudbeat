package compliance.policy.aws_ec2.ensure_public_ingress

import data.compliance.lib.common
import data.compliance.policy.aws_ec2.data_adapter
import data.compliance.policy.aws_ec2.ports
import future.keywords.every
import future.keywords.if

default rule_evaluation := false

finding := result if {
	# filter
	data_adapter.is_nacl_policy

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		input.resource,
	)
}

rule_evaluation if {
	count(data_adapter.ingresses_with_all_ports_open) == 0
	every entry in data_adapter.nacl_ingresses {
		every port in ports.admin_ports {
			not ports.in_range(entry.PortRange.From, entry.PortRange.To, port)
		}
	}
}
