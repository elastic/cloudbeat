package compliance.cis_eks.rules.cis_5_4_5

import data.cis_eks.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as violating_input_use_tcp_instead_of_https
	test.assert_fail(finding) with input as violating_input_use_both_https_and_tcp
	test.assert_fail(finding) with input as violating_input_use_https_only_but_no_certificate
}

test_pass {
	test.assert_pass(finding) with input as valid_input_with_two_https_listeners
}

test_not_evaluated {
	not finding with input as test_data.not_evaluated_input
}

violating_input_use_tcp_instead_of_https := test_data.generate_elb_input_with_two_load_balancers("TCP", null, "TCP", null)

violating_input_use_both_https_and_tcp := test_data.generate_elb_input_with_two_load_balancers("TCP", null, "HTTPS", "adjkdsajad")

violating_input_use_https_only_but_no_certificate := test_data.generate_elb_input_with_two_load_balancers("HTTPS", null, "HTTPS", "adjkdsajad")

valid_input_with_two_https_listeners := test_data.generate_elb_input_with_two_load_balancers("HTTPS", "dsaaasdzx", "HTTPS", "adjkdsajad")
