package compliance.lib.osquery

is_osquery {
    input.osquery
}

is_file {
    is_osquery
    input.osquery.filename
}

filename = file_name {
    is_file
    file_name = input.osquery.filename
}

filemode = file_mode {
    is_file
    file_mode = input.osquery.mode
}

owner_user_id = uid {
    is_file
    uid = input.osquery.uid
}

owner_group_id = gid {
    is_file
    gid = input.osquery.gid
}

file_ownership_match(requierd_uid, requierd_gid) {
    owner_user_id == requierd_uid
    owner_group_id == requierd_gid
} else = false