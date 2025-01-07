package compliance.cis_eks.rules.cis_5_3_1

import data.cis_eks.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as violating_input_no_encryption_configuration
	test.assert_fail(finding) with input as violating_input_empty_encryption_array
	test.assert_fail(finding) with input as violating_input_null_encryption_array
}

test_pass if {
	test.assert_pass(finding) with input as non_violating_input
}

test_not_evaluated if {
	not finding with input as test_data.not_evaluated_input
}

violating_input_no_encryption_configuration := {
	"type": "caas",
	"subType": "aws-eks",
	"resource": {"Cluster": {
		"Arn": "arn:aws:somearn1234:cluster/EKS-demo",
		"CertificateAuthority": {"Data": "some data"},
		"ClientRequestToken": null,
		"CreatedAt": "2021-10-27T11:08:51Z",
		"Endpoint": "https://C07EBEDB096B808626B023DDBF7520DC.gr7.us-east-2.eks.amazonaws.com",
		"Identity": {"Oidc": {"Issuer": "https://oidc.eks.us-east-2.amazonaws.com/id/C07EBdDB096B80AA626B023SS520SS"}},
		"Logging": {"ClusterLogging": [{
			"Enabled": false,
			"Types": [
				"api",
				"audit",
				"authenticator",
				"controllerManager",
				"scheduler",
			],
		}]},
		"Name": "EKS-Elastic-agent-demo",
	}},
}

violating_input_empty_encryption_array := generate_eks_input_with_encryption_config([])

violating_input_null_encryption_array := generate_eks_input_with_encryption_config(null)

non_violating_input := generate_eks_input_with_encryption_config([{
	"Provider": {},
	"Resources": [],
}])

generate_eks_input_with_encryption_config(encryption_config) := result if {
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

	result = test_data.generate_eks_input(logging, encryption_config, true, true, [])
}
