package compliance.lib.common

# set the rule result
calculate_result(evaluation) = "passed" {
    evaluation
} else = "violation"