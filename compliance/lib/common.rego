package compliance.lib.common

# set the rule result
calculate_result(evaluation) = "passed" {
    evaluation
} else = "violation"

file_ownership_match(uid, gid, requierd_uid, requierd_gid) {
    uid == requierd_uid
    gid == requierd_gid
} else = false

# todo: check preformence of regex alternatives
file_permission_match(filemode, user, group, other) {
    pattern = sprintf("0?[0-%d][0-%d][0-%d]", [user, group, other])
    regex.match(pattern, filemode)
} else = false