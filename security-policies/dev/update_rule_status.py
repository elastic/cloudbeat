import os
import common
import yaml

"""
Generates Markdown tables with implemented rules status for all services.
"""

colalign = {
    "Rule Number": "center",
    "Section": "left",
    "Description": "left",
    "Status": "center",
    "Type": "center",
}


def get_implemented_rules(all_rules, benchmark_id):
    """
    Get list of implemented rules in the repository for current service.
    :param all_rules: List of all rules for specified benchmark
    :param benchmark_id: Benchmark ID
    :return: List of implemented rules
    """
    # Set all rules as not implemented by default
    implemented_rules = {str(rule): ":x:" for rule in all_rules}  # ❌

    # Construct path to rules directory for current service
    rules_dir = os.path.join("../bundle", "compliance", f"{benchmark_id}", "rules")

    # Get list of all rule files in the rules directory
    rule_files = os.listdir(rules_dir)

    # Iterate over all rule files
    for rule_file in rule_files:
        # Extract rule number from rule file name
        rule_number = rule_file.removeprefix("cis_").replace("_", ".")

        # Set rule as implemented
        implemented_rules[rule_number] = ":white_check_mark:"  # ✅

    return implemented_rules


def generate_md_table(benchmark_id):
    """
    Generate Markdown table with implemented rules status for current service.
    :param benchmark_id: Benchmark ID
    :return: Markdown table
    """
    rules_data, sections = common.parse_rules_data_from_excel(benchmark_id)

    # Rename "Title" column to "Description"
    rules_data.rename(columns={"Title": "Description"}, inplace=True)

    # Get list of all rules in sheet
    all_rules = rules_data["Rule Number"].to_list()

    # Get list of implemented rules
    implemented_rules = get_implemented_rules(all_rules, benchmark_id)

    # Add implemented rules' column to the data
    for rule, total_status in implemented_rules.items():
        rules_data.loc[rules_data["Rule Number"] == rule, "Status"] = total_status

    rules_data["Section"] = rules_data["Section"].apply(
        lambda section_id: sections[section_id],
    )

    new_order = ["Rule Number", "Section", "Description", "Status", "Type"]
    rules_data = rules_data.reindex(columns=new_order)
    rules_data = rules_data.sort_values("Rule Number")

    rules_data["Rule Number"] = rules_data["Rule Number"].apply(
        get_rule_path,
        benchmark_id=benchmark_id,
        implemented_rules=implemented_rules,
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

    description = f"### {total_implemented}/{total_rules} implemented rules ({total_status:.0%})\n\n"
    description += f"#### Automated rules: {automated_implemented}/{total_automated} ({automated_status:.0%})\n\n"
    description += f"#### Manual rules: {manual_implemented}/{total_manual} ({manual_status:.0%})\n\n"

    return table, description, total_status


def total_rules_status(rules_data):
    """
    Get number of total rules and number of implemented rules.
    :param rules_data: Rules data
    :return: Number of total rules and number of implemented rules
    """
    implemented_rules = rules_data[rules_data["Status"] == ":white_check_mark:"]
    status = len(implemented_rules) / len(rules_data)
    return len(rules_data), len(implemented_rules), status


def automated_rules_status(rules_data):
    """
    Get number of automated rules and number of implemented automated rules.
    :param rules_data: Rules data
    :return: Number of automated rules and number of implemented automated rules
    """
    automated_rules = rules_data[rules_data["Type"] == "Automated"]
    automated_implemented = automated_rules[automated_rules["Status"] == ":white_check_mark:"]
    status = len(automated_implemented) / len(automated_rules)
    return len(automated_rules), len(automated_implemented), status


def manual_rules_status(rules_data):
    """
    Get number of manual rules and number of implemented manual rules.
    :param rules_data: Rules data
    :return: Number of manual rules and number of implemented manual rules
    """
    manual_rules = rules_data[rules_data["Type"] == "Manual"]
    manual_implemented = manual_rules[manual_rules["Status"] == ":white_check_mark:"]
    status = len(manual_implemented) / len(manual_rules)
    return len(manual_rules), len(manual_implemented), status


def get_rule_path(rule, benchmark_id, implemented_rules):
    """
    Get rule path for specified rule and service.
    :param implemented_rules: ‘Implemented’ column values
    :param rule: Rule number
    :param benchmark_id: Benchmark ID
    :return: Rule path in the repository
    """
    if implemented_rules[rule] == ":white_check_mark:":
        return f"[{rule}](bundle/compliance/{benchmark_id}/rules/cis_{rule.replace('.', '_')})"
    else:
        return rule


def update_main_readme_status_badge(percentage, service):
    """
    Update status badge in the main README file.
    :param percentage: Percentage of implemented rules
    :param service: Service name (k8s, eks, aws)
    """

    """
    """
    readme_path = "../README.md"
    badge_api = "https://img.shields.io/badge"
    with open(readme_path, "r+") as f:
        readme = f.readlines()

        if service == "k8s":
            badge = (
                f"[![CIS {service.upper()}]({badge_api}/CIS-Kubernetes%20({percentage:.0f}%25)-326CE5?"
                f"logo=Kubernetes)](RULES.md#k8s-cis-benchmark)\n"
            )
        elif service == "eks":
            badge = (
                f"[![CIS {service.upper()}]({badge_api}/CIS-Amazon%20EKS%20({percentage:.0f}%25)-FF9900?"
                f"logo=Amazon+EKS)](RULES.md#eks-cis-benchmark)\n"
            )
        elif service == "aws":
            badge = (
                f"[![CIS {service.upper()}]({badge_api}/CIS-AWS%20({percentage:.0f}%25)-232F3E?"
                f"logo=Amazon+AWS)](RULES.md#aws-cis-benchmark)\n"
            )
        elif service == "gcp":
            badge = (
                f"[![CIS {service.upper()}]({badge_api}/CIS-GCP%20({percentage:.0f}%25)-4285F4?"
                f"logo=Google+Cloud)](RULES.md#gcp-cis-benchmark)\n"
            )
        elif service == "azure":
            badge = (
                f"[![CIS {service.upper()}]({badge_api}/CIS-AZURE%20({percentage:.0f}%25)-0078D4?"
                f"logo=Microsoft+Azure)](RULES.md#azure-cis-benchmark)\n"
            )

        badge_line = get_badge_line_number(readme, service)
        readme[badge_line] = badge
        f.seek(0)
        f.truncate()
        f.writelines(readme)


def get_badge_line_number(readme, service):
    """
    Get line number of the status badge in the main README file.
    :param readme: Main README file
    :param service: Service name (k8s, eks, aws)
    :return: Line number
    """
    for i, line in enumerate(readme):
        if line.startswith(f"[![CIS {service.upper()}]"):
            return i


if __name__ == "__main__":
    # Set working directory to the dev directory
    os.chdir(os.path.join(common.repo_root.working_dir, "security-policies", "dev"))

    # Write Markdown table to file
    with open("../RULES.md", "w") as f:
        f.write(f"# Rules Status")
        for benchmark_id in common.benchmark.keys():
            print(f"Generating Markdown table for '{benchmark_id}' service")
            f.write(
                f"\n\n## {benchmark_id.removeprefix('cis_').upper()} CIS Benchmark\n\n",
            )
            table, description, percentage = generate_md_table(benchmark_id)
            f.write(description)
            f.write(table)
            update_main_readme_status_badge(
                percentage * 100,
                benchmark_id.removeprefix("cis_"),
            )
        f.write("\n")
