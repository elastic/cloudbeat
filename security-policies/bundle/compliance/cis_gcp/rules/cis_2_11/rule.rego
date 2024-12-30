package compliance.cis_gcp.rules.cis_2_11

import data.compliance.policy.gcp.monitoring.ensure_log_metric_and_alarm_exists as audit

pattern := `protoPayload.methodName="cloudsql.instances.update"`

finding := audit.finding(pattern)
