package compliance.cis.rules.cis_1_1_15

import data.compliance.lib.data_adapter
import data.compliance.lib.common
import data.compliance.cis_k8s

# Ensure that the scheduler.conf file permissions are set to 644 or more restrictive (Automated)
finding = result {
    data_adapter.filename == "scheduler.conf"
    filemode := data_adapter.filemode
    rule_evaluation := common.file_permission_match(filemode, 6, 4, 4)

    # set result
    result := {
        "evaluation" : common.calculate_result(rule_evaluation),
        "evidence" : { "filemode" : filemode },
        "rule_name" : "Ensure that the scheduler.conf file permissions are set to 644 or more restrictive",
        "tags" : array.concat(cis_k8s.default_tags, ["CIS 1.1.15"])
    }
}