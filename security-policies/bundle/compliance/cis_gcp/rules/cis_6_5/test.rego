package compliance.cis_gcp.rules.cis_6_5

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

type := "cloud-database"

subtype := "gcp-sqladmin-instance"

test_violation if {
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

test_pass if {
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

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
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
