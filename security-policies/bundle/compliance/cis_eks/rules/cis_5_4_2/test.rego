package compliance.cis_eks.rules.cis_5_4_2

import data.cis_eks.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as violating_input_private_access_disabled
	test.assert_fail(finding) with input as violating_input_public_access_enabled
	test.assert_fail(finding) with input as violating_input_private_access_enabled_but_public_access_enabled_and_invalid_filter
	test.assert_fail(finding) with input as violating_input_private_access_enabled_but_public_access_enabled_with_valid_filter
}

test_pass if {
	test.assert_pass(finding) with input as non_violating_input
	test.assert_pass(finding) with input as non_violating_input_public_disabled_and_invalid_filter
}

test_not_evaluated if {
	not finding with input as test_data.not_evaluated_input
}

violating_input_private_access_disabled := test_data.generate_eks_input_with_vpc_config(false, true, ["132.1.50.0/0"])

violating_input_public_access_enabled := test_data.generate_eks_input_with_vpc_config(false, true, ["0.0.0.0/0"])

violating_input_private_access_enabled_but_public_access_enabled_and_invalid_filter := test_data.generate_eks_input_with_vpc_config(true, true, ["0.0.0.0/0"])

violating_input_private_access_enabled_but_public_access_enabled_with_valid_filter := test_data.generate_eks_input_with_vpc_config(true, true, ["203.0.113.5/32"])

non_violating_input := test_data.generate_eks_input_with_vpc_config(true, false, ["0.0.0.0/0"])

non_violating_input_public_disabled_and_invalid_filter := test_data.generate_eks_input_with_vpc_config(true, false, ["203.0.113.5/32"])
