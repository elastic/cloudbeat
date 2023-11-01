package compliance.cis_gcp.rules.cis_2_1

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

test_violation {
	# fail when no read/write logs are set for project/folder/org level
	eval_fail with input as test_data.generate_policies_asset([{}])

	# fail when DATA_WRITE is missing from project
	eval_fail with input as test_data.generate_policies_asset([{"iam_policy": {"audit_configs": [{
		"audit_log_configs": [{"log_type": 2}, {"log_type": 1}],
		"service": "allServices",
	}]}}])

	# fail when DATA_READ is missing from project
	eval_fail with input as test_data.generate_policies_asset([{"iam_policy": {"audit_configs": [{
		"audit_log_configs": [{"log_type": 3}, {"log_type": 1}],
		"service": "allServices",
	}]}}])

	# fail when ADMIN_READ is missing from project
	eval_fail with input as test_data.generate_policies_asset([{"iam_policy": {"audit_configs": [{
		"audit_log_configs": [{"log_type": 3}, {"log_type": 2}],
		"service": "allServices",
	}]}}])

	# fail when extempted members is not empty
	eval_fail with input as test_data.generate_policies_asset([{"iam_policy": {"audit_configs": [{
		"audit_log_configs": [
			{
				"log_type": 3,
				"exempted_members": ["user:a"],
			},
			{"log_type": 2}, {"log_type": 1},
		],
		"service": "allServices",
	}]}}])

	# fail when "service": "allServices" is missing from project
	eval_fail with input as test_data.generate_policies_asset([{"iam_policy": {"audit_configs": [{"audit_log_configs": [{"log_type": 3}, {"log_type": 2}, {"log_type": 1}]}]}}])

	# fail when DATA_READ and DATA_WRITE aren't set on the same policy
	eval_fail with input as test_data.generate_policies_asset([
		{"iam_policy": {"audit_configs": [{
			"audit_log_configs": [{"log_type": 3}, {"log_type": 1}],
			"service": "allServices",
		}]}},
		{"iam_policy": {"audit_configs": [{
			"audit_log_configs": [{"log_type": 2}, {"log_type": 1}],
			"service": "allServices",
		}]}},
	])
}

test_pass {
	# passes when project has DATA_READ/DATA_WRITE/ADMIN_READ
	# for all services, and with no exempted members
	eval_pass with input as test_data.generate_policies_asset([{"iam_policy": {"audit_configs": [{
		"audit_log_configs": [
			{"log_type": 1},
			{"log_type": 2},
			{"log_type": 3},
		],
		"service": "allServices",
	}]}}])
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
