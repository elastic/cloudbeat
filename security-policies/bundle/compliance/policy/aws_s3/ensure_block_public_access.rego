package compliance.policy.aws_s3.ensure_block_public_access

import data.compliance.lib.assert
import data.compliance.lib.common as lib_common
import data.compliance.policy.aws_s3.data_adapter
import future.keywords.if
import future.keywords.in

public_access_block_config_is_blocked(config) if {
	config.BlockPublicAcls == true
	config.BlockPublicPolicy == true
	config.IgnorePublicAcls == true
	config.RestrictPublicBuckets == true
} else := false

default rule_evaluation := false

# If we got public access block config for both account and bucket
rule_evaluation if {
	not data_adapter.public_access_block_configuration == null
	not data_adapter.account_public_access_block_configuration == null
	assert.some_true([data_adapter.public_access_block_configuration.BlockPublicAcls, data_adapter.account_public_access_block_configuration.BlockPublicAcls])
	assert.some_true([data_adapter.public_access_block_configuration.BlockPublicPolicy, data_adapter.account_public_access_block_configuration.BlockPublicPolicy])
	assert.some_true([data_adapter.public_access_block_configuration.IgnorePublicAcls, data_adapter.account_public_access_block_configuration.IgnorePublicAcls])
	assert.some_true([data_adapter.public_access_block_configuration.RestrictPublicBuckets, data_adapter.account_public_access_block_configuration.RestrictPublicBuckets])
}

# If we got only account-level public access block config
rule_evaluation if {
	not data_adapter.account_public_access_block_configuration == null
	data_adapter.public_access_block_configuration == null
	public_access_block_config_is_blocked(data_adapter.account_public_access_block_configuration)
}

# If we got only bucket-level public access block config
rule_evaluation if {
	not data_adapter.public_access_block_configuration == null
	data_adapter.account_public_access_block_configuration == null
	public_access_block_config_is_blocked(data_adapter.public_access_block_configuration)
}

finding := result if {
	data_adapter.is_s3

	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"PublicAccessBlockConfiguration": data_adapter.public_access_block_configuration, "AccountPublicAccessBlockConfiguration": data_adapter.account_public_access_block_configuration},
	)
}
