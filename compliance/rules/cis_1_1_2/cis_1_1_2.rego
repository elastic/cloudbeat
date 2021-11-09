package compliance.cis.rules.cis_1_1_2

import data.compliance.lib.osquery
import data.compliance.lib.common
import data.compliance.cis


# Ensure that the API server pod specification file ownership is set to root:root
finding = {"evaluation": evaluation, "rule_name": rule_name, "evidence": evidence, "tags": tags} {
    osquery.filename == "kube-apiserver.yaml"
    uid = osquery.owner_user_id
    gid = osquery.owner_group_id
    rule_evaluation := common.file_ownership_match(uid, gid, "root", "root")

    # set result
    evaluation := common.calculate_result(rule_evaluation)
    evidence := [
        {"key": "uid", "value": uid},
        {"key": "gid", "value": gid}
    ]
    rule_name := "Ensure that the API server pod specification file ownership is set to root:root"
    tags := array.concat(cis.tags, ["CIS 1.1.2"])
}