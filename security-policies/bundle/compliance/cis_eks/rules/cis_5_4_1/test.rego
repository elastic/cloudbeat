package compliance.cis_eks.rules.cis_5_4_1

import data.cis_eks.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as violating_input_private_access_disabled
	test.assert_fail(finding) with input as violating_input_public_invalid_filter
	test.assert_fail(finding) with input as violating_input_private_access_disabled_and_public_access_enabled_valid_filter
}

test_pass if {
	test.assert_pass(finding) with input as non_violating_input
	test.assert_pass(finding) with input as valid_input_public_access_disabled_and_private_endpoint_endabled
}

test_not_evaluated if {
	not finding with input as test_data.not_evaluated_input
}

violating_input_private_access_disabled := test_data.generate_eks_input_with_vpc_config(false, true, ["132.1.50.0/0"])

violating_input_public_invalid_filter := test_data.generate_eks_input_with_vpc_config(true, true, ["0.0.0.0/0"])

violating_input_private_access_disabled_and_public_access_enabled_valid_filter := test_data.generate_eks_input_with_vpc_config(false, true, ["132.1.50.0/0"])

valid_input_public_access_disabled_and_private_endpoint_endabled := test_data.generate_eks_input_with_vpc_config(true, false, ["0.0.0.0/0"])

non_violating_input := test_data.generate_eks_input_with_vpc_config(true, true, ["203.0.113.5/32"])
