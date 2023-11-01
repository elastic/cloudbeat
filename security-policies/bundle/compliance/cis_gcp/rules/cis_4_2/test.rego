package compliance.cis_gcp.rules.cis_4_2

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as rule_input({"name": "test-instance", "serviceAccounts": [{"email": "legit-sa@developer.gserviceaccount.com"}, {"email": "12345678-compute@developer.gserviceaccount.com", "scopes": ["https://www.googleapis.com/auth/cloud-platform"]}]})
}

test_pass {
	eval_pass with input as rule_input({"name": "gke-node"})
	eval_pass with input as rule_input({"name": "gke-node", "serviceAccounts": [{"email": "12345678-compute@developer.gserviceaccount.com"}]})
	eval_pass with input as rule_input({"name": "test-instance", "serviceAccounts": [{"email": "12345678-compute@developer.gserviceaccount.com", "scopes": ["https://www.googleapis.com/auth/compute"]}]})
	eval_pass with input as rule_input({"name": "test-instance", "serviceAccounts": [
		{"email": "test-sa@developer.gserviceaccount.com", "scopes": ["https://www.googleapis.com/auth/cloud-platform"]},
		{"email": "test2-sa@developer.gserviceaccount.com", "scopes": ["https://www.googleapis.com/auth/cloud-platform"]},
	]})
}

test_not_evaluated {
	not_eval with input as test_data.not_eval_resource
}

rule_input(info) = test_data.generate_compute_resource("gcp-compute-instance", info)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
