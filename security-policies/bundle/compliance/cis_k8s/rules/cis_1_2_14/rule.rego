package compliance.cis_k8s.rules.cis_1_2_14

import data.compliance.policy.process.ensure_arguments_not_contain_value_appropriate as audit

finding := audit.finding("--disable-admission-plugins", "ServiceAccount")
