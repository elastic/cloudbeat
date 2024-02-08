import importlib
import importlib.util
import os
import sys
from pkgutil import iter_modules

import common

"""
Generates Markdown tables with implemented rules status for all services.
"""

table_of_contents = """
## Table of Contents\n
- [Kubernetes CIS Benchmark](#k8s-cis-benchmark)
- [Amazon EKS CIS Benchmark](#eks-cis-benchmark)
- [Amazon AWS CIS Benchmark](#aws-cis-benchmark)
- [Google Cloud CIS Benchmark](#gcp-cis-benchmark)
- [Microsoft Azure CIS Benchmark](#azure-cis-benchmark)"""

colalign = {
    "Rule Number": "center",
    "Section": "left",
    "Description": "left",
    "Status": "center",
    "Type": "center",
}

test_cases_variable_name = "test_cases"
test_cases_module_name = "product.tests.data"


def get_integration_test_cases(benchmark_id):
    """
    Given a benchmark_id looks up all the submodules, and retrieve the map declared as test_cases
    @param benchmark_id: benchmark to look for integration tests
    :return: Dictionary of test cases
    """
    root_module_name = f"{test_cases_module_name}.{get_provider(benchmark_id)}"
    if importlib.util.find_spec(root_module_name) is None:
        # No test data found
        return dict()

    root_module = importlib.import_module(root_module_name)
    modules = iter_modules(root_module.__path__)
    tcs = dict()
    for module in modules:
        submodule = importlib.import_module("." + module.name, root_module.__name__)

        if not hasattr(submodule, test_cases_variable_name):
            print("Could not find test_cases in ", root_module.__name__, module.name)
            continue

        test_cases = getattr(submodule, test_cases_variable_name)

        for _, tc in test_cases.items():
            if (
                test_cases_module_name not in type(tc).__module__
                or getattr(tc, "rule_tag") is None
            ):
                continue

            rule_number = tc.rule_tag.replace("CIS ", "")

            if rule_number not in tcs:
                tcs[rule_number] = {
                    "passed": False,
                    "failed": False,
                }

            tcs[rule_number][tc.expected] = True

    return tcs


def generate_integration_test_cel(rule, integration_tests, rule_implementation_status):
    """
    Get formatted integration test status
    @param rule: which rule to look for
    @param integration_tests: all the integration tests
    @param rule_implementation_status: implemented rules status
    :return: pretty string with Passed Failed status
    """
    if rule_implementation_status[rule] != common.positive_emoji:
        return "N/A"

    passed = failed = False
    if rule in integration_tests:
        passed = integration_tests[rule].get("passed", False)
        failed = integration_tests[rule].get("failed", False)
    return f"""Passed {common.status_emoji(passed)} / Failed {common.status_emoji(failed)}"""


def get_implemented_rules(all_rules, benchmark_id):
    """
    Get list of implemented rules in the repository for current service.
    @param benchmark_id: Benchmark ID
    @param all_rules: List of all rules for specified benchmark
    :return: List of implemented rules
    """
    # Set all rules as not implemented by default
    implemented_rules = {str(rule): common.negative_emoji for rule in all_rules}

    # Construct path to rules directory for current service
    rules_dir = os.path.join("../bundle", "compliance", f"{benchmark_id}", "rules")

    # Get list of all rule files in the rules directory
    rule_files = os.listdir(rules_dir)

    # Iterate over all rule files
    for rule_file in rule_files:
        # Extract rule number from rule file name
        rule_number = rule_file.removeprefix("cis_").replace("_", ".")

        # Set rule as implemented
        implemented_rules[rule_number] = common.positive_emoji

    return implemented_rules


def generate_md_table(benchmark_id):
    """
    Generate Markdown table with implemented rules status for current service.
    @param benchmark_id: Benchmark ID
    :return: Markdown table
    """
    rules_data, sections = common.parse_rules_data_from_excel(benchmark_id)

    # Rename "Title" column to "Description"
    rules_data.rename(columns={"Title": "Description"}, inplace=True)

    # Get list of all rules in sheet
    all_rules = rules_data["Rule Number"].to_list()

    # Get list of rule implementation status
    rule_implementation_status = get_implemented_rules(all_rules, benchmark_id)

    # Get integration tests for benchmark
    test_cases = get_integration_test_cases(benchmark_id)

    # Add implemented rules' and Integration Tests column to the data
    for rule, total_status in rule_implementation_status.items():
        rules_data.loc[rules_data["Rule Number"] == rule, "Status"] = total_status
        rules_data.loc[rules_data["Rule Number"] == rule, "Integration Tests"] = (
            generate_integration_test_cel(
                rule=rule,
                integration_tests=test_cases,
                rule_implementation_status=rule_implementation_status,
            )
        )

    rules_data["Section"] = rules_data["Section"].apply(
        lambda section_id: sections[section_id],
    )

    new_order = [
        "Rule Number",
        "Section",
        "Description",
        "Status",
        "Integration Tests",
        "Type",
    ]
    rules_data = rules_data.reindex(columns=new_order)
    rules_data = rules_data.sort_values("Rule Number")

    rules_data["Rule Number"] = rules_data["Rule Number"].apply(
        get_rule_path,
        benchmark_id=benchmark_id,
        implemented_rules=rule_implementation_status,
    )

    # Convert DataFrame to Markdown table
    table = rules_data.to_markdown(
        index=False,
        tablefmt="pipe",
        colalign=colalign.values(),
    )

    # Add table title
    total_rules, total_implemented, total_status = total_rules_status(rules_data)
    total_automated, automated_implemented, automated_status = automated_rules_status(
        rules_data,
    )
    total_manual, manual_implemented, manual_status = manual_rules_status(rules_data)
    total_expected_tests, implemented_tests = integration_test_status(
        rule_implementation_status,
        test_cases,
    )
    # Calculate automation tests coverage
    coverage = (
        implemented_tests / total_expected_tests if total_expected_tests > 0 else 0
    )

    description = f"### {total_implemented}/{total_rules} implemented rules ({total_status:.0%})\n\n"
    description += f"#### Automated rules: {automated_implemented}/{total_automated} ({automated_status:.0%})\n\n"
    description += f"#### Manual rules: {manual_implemented}/{total_manual} ({manual_status:.0%})\n\n"
    description += f"#### Integration Tests Coverage: {implemented_tests}/{total_expected_tests} ({coverage:.0%})\n\n"
    total_percentage = total_status * 100

    return table, description, total_percentage


def total_rules_status(rules_data):
    """
    Get number of total rules and number of implemented rules.
    @param rules_data: Rules data
    :return: Number of total rules and number of implemented rules
    """
    implemented_rules = rules_data[rules_data["Status"] == common.positive_emoji]
    status = len(implemented_rules) / len(rules_data)
    return len(rules_data), len(implemented_rules), status


def integration_test_status(rule_implementation_status, test_cases):
    """
    Calculates the test automation coverage percentage of the implemented rules.
    @param rule_implementation_status: implemented rules status
    @param test_cases: all the test cases
    :return: total expected cases, total test cases and the coverage percentage
    """
    total_expected_test_cases = 0
    total_test_cases = 0
    for rule, status in rule_implementation_status.items():
        if status == common.positive_emoji:
            # each implemented rule should have 2 test cases (passed and failed)
            total_expected_test_cases += 2
        if rule in test_cases:
            total_test_cases += 1 if test_cases[rule]["passed"] else 0
            total_test_cases += 1 if test_cases[rule]["failed"] else 0

    return total_expected_test_cases, total_test_cases


def automated_rules_status(rules_data):
    """
    Get number of automated rules and number of implemented automated rules.
    @param rules_data: Rules data
    :return: Number of automated rules and number of implemented automated rules
    """
    automated_rules = rules_data[rules_data["Type"] == "Automated"]
    automated_implemented = automated_rules[
        automated_rules["Status"] == common.positive_emoji
    ]
    status = len(automated_implemented) / len(automated_rules)
    return len(automated_rules), len(automated_implemented), status


def manual_rules_status(rules_data):
    """
    Get number of manual rules and number of implemented manual rules.
    @param rules_data: Rules data
    :return: Number of manual rules and number of implemented manual rules
    """
    manual_rules = rules_data[rules_data["Type"] == "Manual"]
    manual_implemented = manual_rules[manual_rules["Status"] == common.positive_emoji]
    status = len(manual_implemented) / len(manual_rules)
    return len(manual_rules), len(manual_implemented), status


def get_rule_path(rule, benchmark_id, implemented_rules):
    """
    Get rule path for specified rule and service.
    @param implemented_rules: ‘Implemented’ column values
    @param rule: Rule number
    @param benchmark_id: Benchmark ID
    :return: Rule path in the repository
    """
    if implemented_rules[rule] == common.positive_emoji:
        return f"[{rule}](bundle/compliance/{benchmark_id}/rules/cis_{rule.replace('.', '_')})"
    else:
        return rule


def generate_badge(percentage, service, rules_md_path):
    """
    Generate status badge for specified service.
    @param service: Service name (k8s, eks, aws)
    @param percentage: Percentage of implemented rules
    @param rules_md_path: RULES.md path relative to the README file
    :return: Status badge
    """
    badge_api = "https://img.shields.io/badge"
    match service:
        case "k8s":
            return (
                f"[![CIS {service.upper()}]({badge_api}/CIS-Kubernetes%20({percentage:.0f}%25)-326CE5?"
                f"logo=Kubernetes)]({rules_md_path}#k8s-cis-benchmark)\n"
            )
        case "eks":
            return (
                f"[![CIS {service.upper()}]({badge_api}/CIS-Amazon%20EKS%20({percentage:.0f}%25)-FF9900?"
                f"logo=Amazon+EKS)]({rules_md_path}#eks-cis-benchmark)\n"
            )
        case "aws":
            return (
                f"[![CIS {service.upper()}]({badge_api}/CIS-AWS%20({percentage:.0f}%25)-232F3E?"
                f"logo=Amazon+AWS)]({rules_md_path}#aws-cis-benchmark)\n"
            )
        case "gcp":
            return (
                f"[![CIS {service.upper()}]({badge_api}/CIS-GCP%20({percentage:.0f}%25)-4285F4?"
                f"logo=Google+Cloud)]({rules_md_path}#gcp-cis-benchmark)\n"
            )
        case "azure":
            return (
                f"[![CIS {service.upper()}]({badge_api}/CIS-AZURE%20({percentage:.0f}%25)-0078D4?"
                f"logo=Microsoft+Azure)]({rules_md_path}#azure-cis-benchmark)\n"
            )


def find_rules_md(start_file_path):
    """
    Find the RULES.md file in the repository relative to the starting file.
    @param start_file_path: Path to the starting file (README.md)
    @return: Path to the RULES.md file relative to the starting file
    """
    # Get the directory of the starting file
    start_dir = os.path.dirname(start_file_path)
    # Walk through the directory and subdirectories
    for root, dirs, files in os.walk(start_dir):
        if "RULES.md" in files:
            # Construct the relative path to RULES.md from the starting directory
            return os.path.relpath(os.path.join(root, "RULES.md"), start_dir)


def update_readme_status_badge(percentage, service, readme_path):
    """
    Update status badge in the main README file.
    @param service: Service name (k8s, eks, aws)
    @param percentage: Percentage of implemented rules
    @param readme_path: Path to the README file
    """
    start_file_path = readme_path
    rules_md_path = find_rules_md(start_file_path)

    with open(readme_path, "r+") as f:
        readme = f.readlines()

        badge = generate_badge(percentage, service, rules_md_path)
        badge_line = get_badge_line_number(readme, service)

        readme[badge_line] = badge
        f.seek(0)
        f.truncate()
        f.writelines(readme)


def get_badge_line_number(readme, service):
    """
    Get line number of the status badge in the main README file.
    @param readme: Main README file
    @param service: Service name (k8s, eks, aws)
    :return: Line number
    """
    for i, line in enumerate(readme):
        if line.startswith(f"[![CIS {service.upper()}]"):
            return i


def get_provider(benchmark_id):
    """
    Get provider name from benchmark_id by removing the prefix which is 'cis_'
    """
    return benchmark_id.removeprefix("cis_")


if __name__ == "__main__":
    # Set working directory to the dev directory
    os.chdir(os.path.join(common.repo_root.working_dir, "security-policies", "dev"))

    # Allow to import from tests folder to fetch Integration tests data
    sys.path.append("../../tests/")

    # Write Markdown table to file
    with open("../RULES.md", "w") as f:
        f.write(f"# Rules Status\n")
        table_of_contents = table_of_contents
        f.write(table_of_contents)

        for benchmark_id in common.benchmark.keys():
            print(f"Generating Markdown table for '{benchmark_id}' service")
            benchmark_title = f"{get_provider(benchmark_id).upper()} CIS Benchmark"
            f.write(f"\n\n## {benchmark_title}\n\n")

            table, description, percentage = generate_md_table(benchmark_id)
            f.write(description)
            f.write(f"<details><summary><h3>Full Table 📋</h3></summary>\n\n")
            f.write(table)
            f.write("\n</details>")

            # Update the status badge in Cloudbeat README file
            update_readme_status_badge(
                readme_path="../../README.md",
                percentage=percentage,
                service=get_provider(benchmark_id),
            )

            # Update the status badge in Security-Policies README file
            update_readme_status_badge(
                readme_path="../README.md",
                percentage=percentage,
                service=get_provider(benchmark_id),
            )

        f.write("\n")
