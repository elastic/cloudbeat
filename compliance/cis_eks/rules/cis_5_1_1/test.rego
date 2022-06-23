package compliance.cis_eks.rules.cis_5_1_1

import data.cis_eks.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as violating_input_scan_on_push_disabled
}

test_pass {
	test.assert_pass(finding) with input as valid_input
}

test_not_evaluated {
	not finding with input as test_data.not_evaluated_input
}

violating_input_scan_on_push_disabled = test_data.generate_ecr_input_with_one_repo(false)

valid_input = test_data.generate_ecr_input_with_one_repo(true)
