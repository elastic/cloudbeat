package compliance.cis_k8s.rules.cis_5_1_5

import data.compliance.policy.kube_api.ensure_service_accounts as audit

finding := audit.finding(audit.service_account_default)
