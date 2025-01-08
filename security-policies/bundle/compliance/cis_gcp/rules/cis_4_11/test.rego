package compliance.cis_gcp.rules.cis_4_11

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

eval_machine_type := "https://www.googleapis.com/compute/v1/projects/elastic-security-dev/zones/us-west1-b/machineTypes/n2d-standard-8"

non_eval_machine_type := "https://www.googleapis.com/compute/v1/projects/elastic-security-dev/zones/us-west1-b/machineTypes/n1d-standard-8"

test_violation if {
	eval_fail with input as rule_input({"machineType": eval_machine_type})
	eval_fail with input as rule_input({"confidentialInstanceConfig": {}, "machineType": eval_machine_type})
	eval_fail with input as rule_input({"confidentialInstanceConfig": {"enableConfidentialCompute": false}, "machineType": eval_machine_type})
}

test_pass if {
	eval_pass with input as rule_input({"confidentialInstanceConfig": {"enableConfidentialCompute": true}, "machineType": eval_machine_type})
}

test_not_evaluated if {
	not_eval with input as rule_input({"confidentialInstanceConfig": {"enableConfidentialCompute": true}, "machineType": non_eval_machine_type})
}

rule_input(info) := test_data.generate_compute_resource("gcp-compute-instance", info)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
