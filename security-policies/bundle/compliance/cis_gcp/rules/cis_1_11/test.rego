package compliance.cis_gcp.rules.cis_1_11

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

type := "key-management"

subtype := "gcp-cloudresourcemanager-project"

user_a := "user:a"

admin_role := {
	"role": "roles/cloudkms.admin",
	"members": [user_a],
}

test_violation if {
	# fail when same user (user:a) is both:
	# roles/cloudkms.admin and roles/cloudkms.cryptoKeyEncrypter
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {}, {"bindings": [
		admin_role,
		{
			"role": "roles/cloudkms.cryptoKeyEncrypter",
			"members": [user_a, "user:c"],
		},
	]})

	# fail when same user (user:a) is both:
	# roles/cloudkms.admin and roles/cloudkms.cryptoKeyDecrypter
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {}, {"bindings": [
		admin_role,
		{
			"role": "roles/cloudkms.cryptoKeyDecrypter",
			"members": [user_a, "user:c"],
		},
	]})

	# fail when same user (user:a) is both:
	# roles/cloudkms.admin and roles/cloudkms.cryptoKeyEncrypterDecrypter
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {}, {"bindings": [
		admin_role,
		{
			"role": "roles/cloudkms.cryptoKeyEncrypterDecrypter",
			"members": [user_a, "user:c"],
		},
	]})
}

test_pass if {
	# pass when no user has both roles
	eval_pass with input as test_data.generate_gcp_asset(
		type,
		subtype,
		{},
		{"bindings": [
			admin_role,
			{
				"role": "roles/cloudkms.cryptoKeyEncrypterDecrypter",
				"members": ["user:c", "user:d"],
			},
		]},
	)
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
	not_eval with input as test_data.no_policy_resource
}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
