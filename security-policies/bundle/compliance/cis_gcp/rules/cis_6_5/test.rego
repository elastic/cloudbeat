package compliance.cis_gcp.rules.cis_6_5

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

type := "cloud-database"

subtype := "gcp-sqladmin-instance"

test_violation {
	# fail when authorizedNetworks includes 0.0.0.0/0
	eval_fail with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {"settings": {"ipConfiguration": {"authorizedNetworks": [
			{"value": "10.10.10.10/0"}, # pass
			{"value": "0.0.0.0/0"}, # fail
		]}}}},
		{},
	)
}

test_pass {
	# pass when authorizedNetworks is not defined
	eval_pass with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {"settings": {}}},
		{},
	)

	# pass when authorizedNetworks is not "0.0.0.0/0"
	eval_pass with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {"settings": {"ipConfiguration": {"authorizedNetworks": [{"value": "10.10.10.10/0"}]}}}},
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
