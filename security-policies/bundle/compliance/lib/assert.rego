package compliance.lib.assert

import future.keywords.if

is_true(value) if {
	not is_false(value)
}

# False positive fixed in next version of Regal
# https://github.com/StyraInc/regal/issues/433
# regal ignore:equals-pattern-matching
is_false(value) if {
	value == false
} else := false

all_true(values) if {
	not some_false(values)
}

all_false(values) if {
	not some_true(values)
}

some_false(values) if {
	value := values[_]
	not value
}

some_true(values) if {
	value := values[_]
	value
}

array_is_empty(array) if {
	count(array) == 0
} else := false
