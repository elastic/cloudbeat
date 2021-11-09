package compliance.cis.rules.cis_1_1_11

import data.compliance.lib.osquery
import data.compliance.lib.common
import data.compliance.cis

# Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated)
finding = {"evaluation": evaluation, "rule_name": rule_name, "evidence": evidence, "tags": tags} {
    osquery.filename == "etcd"
    filemode := osquery.filemode
    pattern := "0?(0|1|2|3|4|5|6|7)00"
    rule_evaluation := regex.match(pattern, filemode)

    # set result
    evaluation := common.calculate_result(rule_evaluation)
    evidence := [{ "key": "filemode", "value": filemode }]
    rule_name := "Ensure that the etcd data directory permissions are set to 700 or more restrictive"
    tags := array.concat(cis.tags, ["CIS 1.1.11"])
}