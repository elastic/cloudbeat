package compliance.policy.aws_config.ensure_config_enabled

import data.compliance.lib.common
import data.compliance.policy.aws_config.data_adapter
import future.keywords.every

default rule_evaluation = false

rule_evaluation {
	# every config needs to have at least 1 enabled recorder
	every config in data_adapter.configs {
		recorder := config.recorders[_]
		recorder.ConfigurationRecorder.RecordingGroup.AllSupported == true
		recorder.ConfigurationRecorder.RecordingGroup.IncludeGlobalResourceTypes == true
	}
}

finding = result {
	data_adapter.is_configservice

	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		data_adapter.configs,
	)
}
