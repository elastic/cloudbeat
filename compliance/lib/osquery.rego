package compliance.lib.osquery

is_osquery {
    input.osquery
}

is_file {
    is_osquery
    input.osquery.filename
}

filename = x {
    is_file
    x = input.osquery.filename
}

filemode = x {
    is_file
    x = input.osquery.mode
}

owner_user_id = x {
    is_file
    x = input.osquery.uid
}

owner_group_id = x {
    is_file
    x = input.osquery.gid
}