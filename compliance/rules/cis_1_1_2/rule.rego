package compliance.cis.rules.cis_1_1_2

import data.compliance.lib.data_adapter
import data.compliance.lib.common
import data.compliance.cis


# Ensure that the API server pod specification file ownership is set to root:root
finding = {"evaluation": evaluation, "rule_name": rule_name, "evidence": evidence, "tags": tags} {
    data_adapter.filename == "kube-apiserver.yaml"
    uid = data_adapter.owner_user_id
    gid = data_adapter.owner_group_id
    rule_evaluation := common.file_ownership_match(uid, gid, "root", "root")

    # set result
    evaluation := common.calculate_result(rule_evaluation)
    evidence := {"uid" : uid, "gid" : gid}
    rule_name := "Ensure that the API server pod specification file ownership is set to root:root"
    tags := array.concat(cis.tags, ["CIS 1.1.2"])
}