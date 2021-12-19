package compliance.lib.assert

is_false(value) {
	value == false
} else = false {
	true
}

all_true(values) {
	not some_false(values)
}

some_false(values) {
	value := values[_]
	not value
}
