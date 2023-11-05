package compliance.lib.assert

is_true(value) {
	not is_false(value)
}

# False positive fixed in next version of Regal
# https://github.com/StyraInc/regal/issues/433
# regal ignore:equals-pattern-matching
is_false(value) {
	value == false
} else = false

all_true(values) {
	not some_false(values)
}

all_false(values) {
	not some_true(values)
}

some_false(values) {
	value := values[_]
	not value
}

some_true(values) {
	value := values[_]
	value
}

array_is_empty(array) {
	count(array) == 0
} else = false
