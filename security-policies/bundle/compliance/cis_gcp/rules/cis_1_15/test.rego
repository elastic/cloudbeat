package compliance.cis_gcp.rules.cis_1_15

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

type := "key-management"

subtype := "gcp-apikeys-key"

date_within_last_90_days := time.format(time.add_date(time.now_ns(), 0, 0, -1))

date_before_last_90_days := time.format(time.add_date(time.now_ns(), 0, 0, -91))

test_violation {
	eval_fail with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {"createTime": date_before_last_90_days}},
		{},
	)
}

test_pass {
	eval_pass with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {"createTime": date_within_last_90_days}},
		{},
	)
}

test_not_evaluated {
	not_eval with input as test_data.not_eval_resource
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
