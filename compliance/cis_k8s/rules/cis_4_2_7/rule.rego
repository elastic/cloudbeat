package compliance.cis_k8s.rules.cis_4_2_7

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --make-iptables-util-chains argument is set to true (Automated)

default rule_evaluation = true

process_args := cis_k8s.data_adapter.process_args

rule_evaluation = false {
	common.contains_key_with_value(process_args, "--make-iptables-util-chains", "false")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Checks that the entry for makeIPTablesUtilChains is set to true.
rule_evaluation = false {
	not process_args["--make-iptables-util-chains"]
	data_adapter.process_config.config.makeIPTablesUtilChains == false
}

finding = result {
	# filter
	data_adapter.is_kubelet

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {
			"process_args": process_args,
			"process_config": data_adapter.process_config,
		},
	}
}
