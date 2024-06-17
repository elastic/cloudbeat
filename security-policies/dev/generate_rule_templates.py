import argparse
import json
import os
from pathlib import Path

import common
import yaml

INTEGRATION_RULE_TEMPLATE_DIR = "../../../integrations/packages/cloud_security_posture/kibana/csp_rule_template/"


def generate_rule_templates(
    benchmark: str,
    selected_rules: list,
    rule_template_dir: str,
):
    """
    Generate rule templates from existing rules
    """
    rules_templates = []

    benchmark_rules_dir = f"../bundle/compliance/{benchmark}/rules/"
    rules = os.listdir(benchmark_rules_dir)

    # if no rules are selected, generate all rules
    if not selected_rules:
        selected_rules = rules

    for rule in rules:
        # supporting both cis_1_1_1 and 1.1.1 formats
        if rule not in selected_rules and rule.removeprefix("cis_").replace("_", ".") not in selected_rules:
            continue

        with open(
            os.path.join(benchmark_rules_dir, f"{rule}/data.yaml"),
            "r",
            encoding="utf-8",
        ) as f:
            rule_obj = yaml.safe_load(f.read())["metadata"]
            rule_obj["rego_rule_id"] = rule

            rule_template = migrate_csp_rule_metadata(
                {
                    "id": rule_obj["id"],
                    "type": "csp-rule-template",
                    "attributes": rule_obj,
                },
            )

            rules_templates.append(rule_template)

    # Write templates into file
    print(f"Processed {len(rules_templates)} rules for {benchmark} benchmark")
    save_rule_templates(rules_templates, rule_template_dir)


def save_rule_templates(rule_templates: list[dict], rule_template_dir: str):
    """
    Save rule templates to file
    """
    if not rule_template_dir:
        rule_template_dir = INTEGRATION_RULE_TEMPLATE_DIR

    # ensure path exist, else create it
    Path(rule_template_dir).mkdir(parents=True, exist_ok=True)

    for rule_template in rule_templates:
        with open(
            os.path.join(rule_template_dir, f"{rule_template['id']}.json"),
            "w",
        ) as f:
            json.dump(rule_template, f, indent=4)


def migrate_csp_rule_metadata(doc: dict) -> dict:
    """
    Migrate rule metadata to integration format
    """
    attributes = doc["attributes"]
    print(f"Processing {attributes['benchmark']['rule_number']}")
    metadata = {
        "impact": attributes.pop("impact", None),
        "default_value": attributes.pop("default_value", None),
        "references": attributes.pop("references", None),
        **attributes,
    }
    return {
        **doc,
        "attributes": {
            "metadata": metadata,
        },
        "migrationVersion": {
            "csp-rule-template": "8.7.0",
        },
        "coreMigrationVersion": "8.7.0",
    }


if __name__ == "__main__":
    os.chdir(os.path.join(common.repo_root.working_dir, "security-policies", "dev"))

    parser = argparse.ArgumentParser(
        description="CIS Benchmark Rules Templates Generator CLI",
    )
    parser.add_argument(
        "-b",
        "--benchmark",
        default=common.benchmark.keys(),
        choices=common.benchmark.keys(),
        help="benchmark to be used for the rules template generation (default: all benchmarks). "
        "for example: `--benchmark cis_eks` or `--benchmark cis_eks cis_aws`",
        nargs="+",
    )
    parser.add_argument(
        "-r",
        "--rules",
        default=[],
        help="set of specific rules to be parsed (default: all rules)."
        "for example: `--rules 1.1 1.2` or `--rules cis_1_1 cis_1_2`",
        nargs="+",
    )
    parser.add_argument(
        "-o",
        "--out",
        help=f"output directory for the generated rules templates (default: {INTEGRATION_RULE_TEMPLATE_DIR}).",
        default=INTEGRATION_RULE_TEMPLATE_DIR,
    )
    args = parser.parse_args()

    if type(args.benchmark) is str:
        args.benchmark = [args.benchmark]

    if type(args.rules) is str:
        args.rules = [args.rules]

    for benchmark_id in args.benchmark:
        print(f"### Processing {benchmark_id.replace('_', ' ').upper()}")
        generate_rule_templates(benchmark_id, args.rules, args.out)
