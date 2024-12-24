package compliance.cis_aws.rules.cis_1_19

import data.compliance.cis_aws.data_adapter
import data.compliance.lib.common
import data.lib.test
import future.keywords.if

generate_certificate_resource(certificates) := {
	"subType": "aws-iam-server-certificate",
	"resource": {"certificates": certificates},
}

generate_expiration(expiration) := {"Expiration": expiration}

last_year := common.create_date_from_ns(time.add_date(time.now_ns(), -1, 0, 0))

next_year := common.create_date_from_ns(time.add_date(time.now_ns(), 1, 0, 0))

test_violation if {
	# fails when an expired certificate exists
	eval_fail with input as generate_certificate_resource([generate_expiration(last_year)])
	eval_fail with input as generate_certificate_resource([
		generate_expiration(last_year),
		generate_expiration(next_year),
	])
}

test_pass if {
	# passes when certificates are not expired or when there are none
	eval_pass with input as generate_certificate_resource([])
	eval_pass with input as generate_certificate_resource([generate_expiration(next_year)])
}

test_not_evaluated if {
	not_eval with input as {"subType": "unknown"}
}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
