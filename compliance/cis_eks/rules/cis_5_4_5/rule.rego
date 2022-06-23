package compliance.cis_eks.rules.cis_5_4_5

import data.compliance.cis_eks.data_adapter
import data.compliance.lib.common

default rule_evaluation = true

# Verify that all listeners has an SSL Certificate
rule_evaluation = false {
	input.resource.ListenerDescriptions[_].Listener.SSLCertificateId == null
}

# Verify that all listeners use https protocoal
rule_evaluation = false {
	input.resource.ListenerDescriptions[_].Listener.InstanceProtocol != "HTTPS"
}

# Ensure there Kuberenetes endpoint private access is enabled
finding = result {
	# filter
	data_adapter.is_aws_elb

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {
			"load_balancer_name": input.resource.LoadBalancerName,
			"listener_descriptions": input.resource.ListenerDescriptions,
		},
	}
}
