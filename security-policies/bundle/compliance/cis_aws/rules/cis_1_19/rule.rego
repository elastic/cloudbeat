package compliance.cis_aws.rules.cis_1_19

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.every

default rule_evaluation = false

finding = result {
	data_adapter.is_server_certificate

	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		{"certificates": data_adapter.server_certificates},
	)
}

is_expired(date) {
	then := time.parse_rfc3339_ns(date)
	now := time.now_ns()
	then < now
}

rule_evaluation {
	every certificate in data_adapter.server_certificates {
		not is_expired(certificate.Expiration)
	}
}
