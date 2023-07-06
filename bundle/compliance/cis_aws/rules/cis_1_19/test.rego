package compliance.cis_aws.rules.cis_1_19

import data.compliance.cis_aws.data_adapter
import data.compliance.lib.common
import data.lib.test

generate_certificate_resource(certificates) = {
	"subType": "aws-iam-server-certificate",
	"resource": {"certificates": certificates},
}

generate_expiration(expiration) = {"Expiration": expiration}

last_year = common.create_date_from_ns(time.add_date(time.now_ns(), -1, 0, 0))

next_year = common.create_date_from_ns(time.add_date(time.now_ns(), 1, 0, 0))

test_violation {
	# fails when an expired certificate exists
	eval_fail with input as generate_certificate_resource([generate_expiration(last_year)])
	eval_fail with input as generate_certificate_resource([
		generate_expiration(last_year),
		generate_expiration(next_year),
	])
}

test_pass {
	# passes when certificates are not expired or when there are none
	eval_pass with input as generate_certificate_resource([])
	eval_pass with input as generate_certificate_resource([generate_expiration(next_year)])
}

test_not_evaluated {
	not_eval with input as {"subType": "unknown"}
}

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
