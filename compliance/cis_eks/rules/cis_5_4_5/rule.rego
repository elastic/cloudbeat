package compliance.cis_eks.rules.cis_5_4_5

import data.compliance.cis_eks.data_adapter
import data.compliance.lib.common

default rule_evaluation = true

# Verify that all listeners has SSL Certificate
rule_evaluation = false {
	input.resource.LoadBalancersDescription[_].ListenerDescriptions[_].Listener.SSLCertificateId == null
}

# Verify that all listeners use https protocoal
rule_evaluation = false {
	input.resource.LoadBalancersDescription[_].ListenerDescriptions[_].Listener.InstanceProtocol != "HTTPS"
}

# Verify that all listeners use https protocoal
evidence["none_https_protocol"] = {"instances:": result} {
	result = [elb_name |
		elb_name = input.resource.LoadBalancersDescription[index].LoadBalancerName
		input.resource.LoadBalancersDescription[index].ListenerDescriptions[_].Listener.InstanceProtocol != "HTTPS"
	]
}

evidence["ssl_certificate_missing"] = {"instances:": result} {
	result = [elb_name |
		elb_name = input.resource.LoadBalancersDescription[index].LoadBalancerName
		input.resource.LoadBalancersDescription[index].ListenerDescriptions[_].Listener.SSLCertificateId == null
	]
}

# Ensure there Kuberenetes endpoint private access is enabled
finding = result {
	# filter
	data_adapter.is_aws_elb

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": evidence,
	}
}