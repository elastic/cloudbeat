package compliance.cis.rules.cis_1_1_2

import data.compliance.lib.osquery
import data.compliance.lib.common
import data.compliance.cis


# Ensure that the API server pod specification file ownership is set to root:root
finding = {"evaluation": evaluation, "rule_name": rule_name, "fields": fields, "tags": tags} {
    osquery.filename == "kube-apiserver.yaml"
    rule_evaluation := osquery.file_ownership_match("root", "root")

    # set result
    evaluation := common.calculate_result(rule_evaluation)
    fields := [
        {"key": "uid", "value": osquery.owner_user_id},
        {"key": "gid", "value": osquery.owner_group_id}
    ]
    rule_name := "Ensure that the API server pod specification file ownership is set to root:root"
    tags := array.concat(cis.tags, ["CIS 1.1.2"])
}