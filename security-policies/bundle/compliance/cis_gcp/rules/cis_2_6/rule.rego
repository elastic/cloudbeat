package compliance.cis_gcp.rules.cis_2_6

import data.compliance.policy.gcp.monitoring.ensure_log_metric_and_alarm_exists as audit

pattern := `resource.type="iam_role"
 AND (protoPayload.methodName = "google.iam.admin.v1.CreateRole" OR
 protoPayload.methodName="google.iam.admin.v1.DeleteRole" OR
 protoPayload.methodName="google.iam.admin.v1.UpdateRole")`

finding := audit.finding(pattern)
