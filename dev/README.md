# Dev tools

## Rules Assets Generators

### `generate_rule_metadata.py`

This script generates the metadata for a rule. It is used to generate the metadata for the rules in the `rules` directory.

**Usage:**

```shell
poetry run python dev/generate_rule_metadata.py --benchmark <benchmark_id> --rules <selected rules>
```

**Example 1** - Generate all rules metadata from all benchmarks:

```shell
poetry run python dev/generate_rule_metadata.py
```

**Example 2** - Generate two specific rules metadata from CIS AWS:

```shell
poetry run python dev/generate_rule_metadata.py --benchmark cis_aws --rules "1.8" "1.9"
```

### Limitations

The script currently has the following limitations:

- It only works with Excel spreadsheets as input.
- It does not generate default values for rules. Default values must be added manually if they are not present in the input spreadsheet.
- Rules rego implementation is required before running the script. The script will fail if the rego implementation is not present.
