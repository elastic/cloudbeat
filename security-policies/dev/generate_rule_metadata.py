import argparse
import os
import uuid
from dataclasses import asdict, dataclass

import common
import pandas as pd
from ruamel.yaml import YAML

yml = YAML()

KSPM_POSTURE_TYPE = "kspm"
CSPM_POSTURE_TYPE = "cspm"


@dataclass
class Benchmark:
    name: str
    version: str
    id: str
    rule_number: str
    posture_type: str


@dataclass
class Rule:
    id: str
    name: str
    profile_applicability: str
    description: str
    rationale: str
    audit: str
    remediation: str
    impact: str
    default_value: str
    references: str
    section: str
    version: str
    tags: list[str]
    benchmark: Benchmark


selected_columns_map = {
    "cis_k8s": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title": "Title",
        "Description": "description",
        "Rational Statement": "rationale",
        "Audit Procedure": "audit",
        "Remediation Procedure": "remediation",
        "Impact Statement": "impact",
        # "": "default_value", # todo: talk with CIS team to add this column to the excel
        "references": "references",
        "Assessment Status": "type",
    },
    "cis_eks": {
        "section #": "Section",
        "recommendation #": "Rule Number",
        "title": "Title",
        "description": "description",
        "rationale statement": "rationale",
        "audit procedure": "audit",
        "remediation procedure": "remediation",
        "impact statement": "impact",
        # "": "default_value", # todo: talk with CIS team to add this column to the excel
        "references": "references",
        "scoring status": "type",
    },
    "cis_aws": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title": "Title",
        "Description": "description",
        "Rational Statement": "rationale",
        "Audit Procedure": "audit",
        "Remediation Procedure": "remediation",
        "Impact Statement": "impact",
        # "": "default_value", # todo: talk with CIS team to add this column to the excel
        "References": "references",
        "Assessment Status": "type",
    },
    "cis_gcp": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title": "Title",
        "Description": "description",
        "Rationale Statement": "rationale",
        "Audit Procedure": "audit",
        "Remediation Procedure": "remediation",
        "Impact Statement": "impact",
        # "": "default_value", # todo: talk with CIS team to add this column to the excel
        "References": "references",
        "Assessment Status": "type",
    },
    "cis_azure": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title": "Title",
        "Description": "description",
        "Rationale Statement": "rationale",
        "Audit Procedure": "audit",
        "Remediation Procedure": "remediation",
        "Impact Statement": "impact",
        # "": "default_value", # todo: talk with CIS team to add this column to the excel
        "References": "references",
        "Assessment Status": "type",
    },
}

benchmark_to_posture_type = {
    "cis_k8s": KSPM_POSTURE_TYPE,
    "cis_eks": KSPM_POSTURE_TYPE,
    "cis_aws": CSPM_POSTURE_TYPE,
    "cis_gcp": CSPM_POSTURE_TYPE,
    "cis_azure": CSPM_POSTURE_TYPE,
}


def parse_refs(refs: str):
    """
    Parse references - they are split by `:` which is the worst token possible for urls...
    """
    if refs != "":
        ref = [f"http{ref}" for ref in refs.split(":http") if ref]
        ref[0] = ref[0].removeprefix("http")
        return "\n".join(f"{i + 1}. {s}" for i, s in enumerate(ref))

    return refs


def read_existing_default_value(rule_number, benchmark_id):
    """
    Read default value from existing rule (The excel file doesn't contain default values)
    :param rule_number: Rule number
    :param benchmark_id: Benchmark ID
    :return: Default value
    """
    rule_dir = os.path.join(
        common.rules_dir,
        f"{benchmark_id}/rules",
        f"cis_{rule_number.replace('.', '_')}",
    )
    try:
        with open(os.path.join(rule_dir, "data.yaml"), "r") as f:
            data = yml.load(f)
            default_value = data["metadata"]["default_value"]
            if default_value is None or default_value == "":
                print(
                    f"{benchmark_id}/{rule_number} is missing default value - please make sure to add it manually",
                )
                return ""
            return data["metadata"]["default_value"]
    except FileNotFoundError:
        print(f"Rule implementation for {benchmark_id}/{rule_number} is missing")
        return ""


def generate_rule_benchmark_metadata(benchmark_id: str, rule_number: str):
    """
    Generate benchmark metadata for rules
    :param benchmark_id: Benchmark ID
    :param rule_number: Rule number
    """
    return Benchmark(
        name=common.benchmark[benchmark_id].split("Benchmark")[0].replace("_", " ").removesuffix(" "),
        version=common.benchmark[benchmark_id].split("Benchmark")[1].removeprefix("_").removesuffix(".xlsx"),
        id=f"{benchmark_id}",
        rule_number=rule_number,
        posture_type=benchmark_to_posture_type[benchmark_id],
    )


def replace_nan_with_empty_string(data: pd.DataFrame):
    """
    Replace NaN values with empty strings (they are represented as `nan` in the Excel for some reason)
    """
    return data.replace("nan", "")


def rule_is_implemented(rule_number: str, benchmark_id: str):
    """
    Check if rule was implemented
    :param rule_number: Rule number
    :param benchmark_id: Benchmark ID
    :return: True if rule was implemented, False otherwise
    """
    rule_path = os.path.join(
        common.rules_dir,
        f"{benchmark_id}/rules",
        f"cis_{rule_number.replace('.', '_')}",
    )
    return os.path.isdir(rule_path)


def generate_metadata(benchmark_id: str, raw_data: pd.DataFrame, sections: dict):
    """
    Generate metadata for rules
    :param benchmark_id: Benchmark ID
    :param raw_data: ‘Raw’ data from the spreadsheet
    :param sections: Section metadata
    :return: List of Rule objects
    """
    normalized_data = replace_nan_with_empty_string(raw_data)
    metadata = []
    benchmark_tag = benchmark_id.removeprefix("cis_").upper() if benchmark_id != "cis_k8s" else f"Kubernetes"
    for rule in normalized_data.to_dict(orient="records"):
        # Check if rule was implemented
        if not rule_is_implemented(rule["Rule Number"], benchmark_id):
            continue

        benchmark_metadata = generate_rule_benchmark_metadata(
            benchmark_id,
            rule["Rule Number"],
        )
        r = Rule(
            id=str(
                uuid.uuid5(
                    uuid.NAMESPACE_DNS,
                    f"{benchmark_metadata.name} {rule['Title']} {rule['Rule Number']}",
                ),
            ),
            name=rule["Title"],
            profile_applicability=f"* {rule['profile_applicability']}",
            description=common.fix_code_blocks(rule["description"], rule['Rule Number'], benchmark_id),
            rationale=common.fix_code_blocks(rule.get("rationale", ""), rule['Rule Number'], benchmark_id),
            audit=common.fix_code_blocks(rule.get("audit", ""), rule['Rule Number'], benchmark_id),
            remediation=common.fix_code_blocks(rule.get("remediation", ""), rule['Rule Number'], benchmark_id),
            impact=rule.get("impact", ""),
            default_value=rule.get(
                "default_value",
                read_existing_default_value(rule["Rule Number"], benchmark_id),
            ),
            references=parse_refs(rule.get("references", "")),
            section=sections[rule["Section"]],
            tags=[
                "CIS",
                benchmark_tag,
                f"CIS {rule['Rule Number']}",
                sections[rule["Section"]],
            ],
            version="1.0",
            benchmark=benchmark_metadata,
        )
        metadata.append(r)

    return metadata


def save_metadata(metadata: list[Rule], benchmark_id):
    """
    Save metadata to file
    :param metadata: List of Rule objects
    :param benchmark_id: Benchmark ID
    :return: None
    """
    for rule in metadata:
        rule_package = f"cis_{rule.benchmark.rule_number.replace('.', '_')}"
        rule_dir = os.path.join(common.rules_dir, f"{benchmark_id}/rules", rule_package)
        try:
            with open(os.path.join(rule_dir, "data.yaml"), "w+") as f:
                yml.dump({"metadata": common.apply_pss_recursively(asdict(rule))}, f)

        except FileNotFoundError:
            continue  # ignore rules that are not implemented


if __name__ == "__main__":
    os.chdir(os.path.join(common.repo_root.working_dir, "security-policies", "dev"))

    parser = argparse.ArgumentParser(
        description="CIS Benchmark parser CLI",
    )
    parser.add_argument(
        "-b",
        "--benchmark",
        default=common.benchmark.keys(),
        choices=common.benchmark.keys(),
        help="benchmark to be used for the rules metadata generation (default: all benchmarks). "
        "for example: `--benchmark cis_eks` or `--benchmark cis_eks cis_aws`",
        nargs="+",
    )
    parser.add_argument(
        "-r",
        "--rules",
        help="set of specific rules to be parsed (default: all rules).",
        nargs="+",
    )
    args = parser.parse_args()

    if type(args.benchmark) is str:
        args.benchmark = [args.benchmark]

    for benchmark_id in args.benchmark:
        print(f"### Processing {benchmark_id.replace('_', ' ').upper()}")

        # Parse Excel data
        raw_data, sections = common.parse_rules_data_from_excel(
            selected_columns=selected_columns_map,
            benchmark_id=benchmark_id,
            selected_rules=args.rules,
        )

        metadata = generate_metadata(benchmark_id, raw_data, sections)
        save_metadata(metadata, benchmark_id)
