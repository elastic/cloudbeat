package compliance.cis_gcp.rules.cis_3_8

import data.cis_gcp.test_data
import future.keywords.if

import data.compliance.policy.gcp.data_adapter
import data.lib.test

type := "cloud-compute"

subtype := "gcp-compute-subnetwork"

# regal ignore:rule-length
test_violation if {
	# fail when enableFlowLogs is missing
	eval_fail with input as test_data.generate_gcp_asset(
		type,
		subtype,
		{"data": {}},
		null,
	)

	# fail when enableFlowLogs is set to false
	eval_fail with input as test_data.generate_gcp_asset(
		type,
		subtype,
		{"data": {"enableFlowLogs": false}},
		null,
	)

	# fail when aggregationInterval is not set to INTERVAL_5_SEC
	eval_fail with input as test_data.generate_gcp_asset(
		type,
		subtype,
		{"data": {
			"purpose": "PRIVATE",
			"enableFlowLogs": true,
			"logConfig": {"aggregationInterval": "INTERVAL_15_SEC"},
		}},
		null,
	)

	# fail when metadata is not set to INCLUDE_ALL_METADATA
	eval_fail with input as test_data.generate_gcp_asset(
		type,
		subtype,
		{"data": {
			"purpose": "PRIVATE",
			"enableFlowLogs": true,
			"logConfig": {
				"aggregationInterval": "INTERVAL_5_SEC",
				"metadata": "foo",
			},
		}},
		null,
	)

	# fail when flowSampling is not set to 1
	eval_fail with input as test_data.generate_gcp_asset(
		type,
		subtype,
		{"data": {
			"purpose": "PRIVATE",
			"enableFlowLogs": true,
			"logConfig": {
				"aggregationInterval": "INTERVAL_5_SEC",
				"metadata": "INCLUDE_ALL_METADATA",
				"flowSampling": 0.5,
			},
		}},
		null,
	)

	# fail when logConfig.enable is not set to true
	eval_fail with input as test_data.generate_gcp_asset(
		type,
		subtype,
		{"data": {
			"purpose": "PRIVATE",
			"enableFlowLogs": true,
			"logConfig": {
				"aggregationInterval": "INTERVAL_5_SEC",
				"metadata": "INCLUDE_ALL_METADATA",
				"flowSampling": 1,
				"enable": false,
			},
		}},
		null,
	)
}

test_pass if {
	eval_pass with input as test_data.generate_gcp_asset(
		type,
		subtype,
		{"data": {
			"purpose": "PRIVATE",
			"enableFlowLogs": true,
			"logConfig": {
				"aggregationInterval": "INTERVAL_5_SEC",
				"metadata": "INCLUDE_ALL_METADATA",
				"flowSampling": 1,
				"enable": true,
			},
		}},
		null,
	)
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
	not_eval with input as test_data.generate_gcp_asset(
		type,
		subtype,
		{"data": {"purpose": "INTERNAL_HTTPS_LOAD_BALANCER"}},
		null,
	)
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
