package compliance.cis_eks.rules.cis_2_1_1

import data.cis_eks.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as violating_input_all_logs_disabled
	test.assert_fail(finding) with input as violating_input_some_disabled
}

test_pass {
	test.assert_pass(finding) with input as non_violating_input
}

test_not_evaluated {
	not finding with input as test_data.not_evaluated_input
}

violating_input_all_logs_disabled = result {
	logging = {"ClusterLogging": [{
		"Enabled": false,
		"Types": [
			"api",
			"audit",
			"authenticator",
			"controllerManager",
			"scheduler",
		],
	}]}

	result = generate_eks_input_with_log(logging)
}

violating_input_some_disabled = result {
	logging = {"ClusterLogging": [
		{
			"Enabled": false,
			"Types": [
				"authenticator",
				"controllerManager",
				"scheduler",
			],
		},
		{
			"Enabled": true,
			"Types": [
				"api",
				"audit",
			],
		},
	]}

	result = generate_eks_input_with_log(logging)
}

non_violating_input = result {
	logging = {"ClusterLogging": [{
		"Enabled": true,
		"Types": [
			"api",
			"audit",
			"authenticator",
			"controllerManager",
			"scheduler",
		],
	}]}

	result = generate_eks_input_with_log(logging)
}

generate_eks_input_with_log(logging) = result {
	encryption_config = {"EncryptionConfig : null"}
	result = test_data.generate_eks_input(logging, encryption_config, true, true, [])
}
