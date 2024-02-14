package compliance.policy.aws_elb.ensure_certificates

import future.keywords.every
import future.keywords.if

import data.compliance.lib.common as lib_common
import data.compliance.policy.aws_elb.data_adapter

default rule_evaluation := false

rule_evaluation if {
	all_https
	not any_null_certificate
}

# Verify that all listeners has an SSL Certificate
any_null_certificate if {
	data_adapter.listener_descriptions[_].Listener.SSLCertificateId == null
}

# Verify that all listeners use https protocoal
all_https if {
	every description in data_adapter.listener_descriptions {
		description.Listener.Protocol == "HTTPS"
	}
}

finding := result if {
	data_adapter.is_aws_elb

	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{
			"load_balancer_name": data_adapter.load_balancer_name,
			"listener_descriptions": data_adapter.listener_descriptions,
		},
	)
}
