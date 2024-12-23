package compliance.cis_gcp.rules.cis_2_10

import data.compliance.policy.gcp.monitoring.ensure_log_metric_and_alarm_exists as audit

pattern := `resource.type="gcs_bucket"
 AND protoPayload.methodName="storage.setIamPermissions"`

finding := audit.finding(pattern)
