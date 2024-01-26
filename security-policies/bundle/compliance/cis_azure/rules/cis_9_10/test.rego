package compliance.cis_azure.rules.cis_9_10

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

kind := "azure-web-site"

test_violation if {
	eval_fail with input as test_data.generate_azure_asset_with_ext(kind, {}, {})
	eval_fail with input as test_data.generate_azure_asset_with_ext(kind, {"hey": {}}, {"authSettings": {}})
	eval_fail with input as test_data.generate_azure_asset_with_ext(kind, {}, {"siteConfig": null})
	eval_fail with input as test_data.generate_azure_asset_with_ext(kind, {}, {"siteConfig": {}})
	eval_fail with input as test_data.generate_azure_asset_with_ext(kind, {}, {"siteConfig": {"properties": {}}})
	eval_fail with input as test_data.generate_azure_asset_with_ext(kind, {}, {"siteConfig": {"properties": {"FtpsState": null}}})
	eval_fail with input as test_data.generate_azure_asset_with_ext(kind, {}, {"siteConfig": {"properties": {"FtpsState": "AllAllowed"}}})
}

test_pass if {
	eval_pass with input as test_data.generate_azure_asset_with_ext(kind, {}, {"siteConfig": {"properties": {"FtpsState": "FtpsOnly"}}})
	eval_pass with input as test_data.generate_azure_asset_with_ext(kind, {}, {"siteConfig": {"properties": {"FtpsState": "Disabled"}}})
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
