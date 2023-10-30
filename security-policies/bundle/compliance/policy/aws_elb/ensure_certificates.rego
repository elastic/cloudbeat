package compliance.policy.aws_elb.ensure_certificates

import data.compliance.lib.common as lib_common
import data.compliance.policy.aws_elb.data_adapter

default rule_evaluation = true

# Verify that all listeners has an SSL Certificate
rule_evaluation = false {
	data_adapter.listener_descriptions[_].Listener.SSLCertificateId == null
}

# Verify that all listeners use https protocoal
rule_evaluation = false {
	data_adapter.listener_descriptions[_].Listener.InstanceProtocol != "HTTPS"
}

finding = result {
	data_adapter.is_aws_elb

	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{
			"load_balancer_name": data_adapter.load_balancer_name,
			"listener_descriptions": data_adapter.listener_descriptions,
		},
	)
}
