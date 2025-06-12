# Dev tools

## Rules Assets Generators

### Prerequisites

We use Poetry to manage our python dependencies. For more details, see [the poetry docs](https://python-poetry.org/docs/).

1. Install poetry (follow [the instructions](https://python-poetry.org/docs/#installation))

2. Install poetry env (follow installing pre-existing project [instructions](https://python-poetry.org/docs/basic-usage/#initialising-a-pre-existing-project))

### Generate Rules Metadata

`generate_rule_metadata.py` generates the metadata for a rule.
It is used to generate the metadata for the rules in the `rules` directory (`data.yaml`).

**Usage:**

From the root dir you can run the following example to generate selected benchmark rules metadata:

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

### Generate Rule Templates

`generate_rule_templates.py` generate the rule templates that will show in our Kibana plug-in (csp-rules).

**Usage:**

From the root dir you can run the following example to generate selected benchmark rules templates

```shell
poetry run python dev/generate_rule_templates.py --benchmark <benchmark_id> --rules <selected rules>
```

**Example 1** - Generate all rules templates from all benchmarks:

```shell
poetry run python dev/generate_rule_templates.py
```

**Example 2** - Generate two specific rules templates from CIS AWS:

```shell
poetry run python dev/generate_rule_templates.py --benchmark cis_aws --rules "1.8" "1.9"
```

**Example 3** - Generate two specific rules templates from CIS AWS and save them in a different directory (relative to `./dev`):

```shell
poetry run python dev/generate_rule_templates.py --benchmark cis_aws --rules "1.8" "1.9"  --out "./rules_templates"
```

> **Note**
> Default output path is the csp integration templates' directory, assuming both repos are sharing the same directory,
> i.e, `../../integrations/packages/cloud_security_posture/kibana/csp_rule_template/`
> This can be configured with the `--out` parameter.
