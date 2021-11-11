package compliance.cis.rules.cis_1_1_3

import data.compliance.lib.data_adapter
import data.compliance.lib.common
import data.compliance.cis_k8s

# Ensure that the API server pod specification file permissions are set to 644 or more restrictive
finding = result {
    data_adapter.filename == "kube-controller-manager.yaml"
    filemode := data_adapter.filemode
    rule_evaluation := common.file_permission_match(filemode, 6, 4, 4)

    # set result
    result := {
        "evaluation" : common.calculate_result(rule_evaluation),
        "evidence" : { "filemode" : filemode },
        "rule_name" : "Ensure that the controller manager pod specification file permissions are set to 644 or more restrictive",
        "tags" : array.concat(cis_k8s.default_tags, ["CIS 1.1.3"])
    }
}