package compliance.cis_gcp.rules.cis_4_11

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

eval_machine_type = "https://www.googleapis.com/compute/v1/projects/elastic-security-dev/zones/us-west1-b/machineTypes/n2d-standard-8"

non_eval_machine_type = "https://www.googleapis.com/compute/v1/projects/elastic-security-dev/zones/us-west1-b/machineTypes/n1d-standard-8"

test_violation {
	eval_fail with input as rule_input({"machineType": eval_machine_type})
	eval_fail with input as rule_input({"confidentialInstanceConfig": {}, "machineType": eval_machine_type})
	eval_fail with input as rule_input({"confidentialInstanceConfig": {"enableConfidentialCompute": false}, "machineType": eval_machine_type})
}

test_pass {
	eval_pass with input as rule_input({"confidentialInstanceConfig": {"enableConfidentialCompute": true}, "machineType": eval_machine_type})
}

test_not_evaluated {
	not_eval with input as rule_input({"confidentialInstanceConfig": {"enableConfidentialCompute": true}, "machineType": non_eval_machine_type})
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
