import os
import common

"""
Generates Markdown tables with implemented rules status for all services.
"""


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
    rules_data.rename(columns={'Title': 'Description'}, inplace=True)

    # Get list of all rules in sheet
    all_rules = rules_data["Rule Number"].to_list()

    # Get list of implemented rules
    implemented_rules = get_implemented_rules(all_rules, benchmark_id)

    # Add implemented rules' column to the data
    for rule, status in implemented_rules.items():
        rules_data.loc[rules_data["Rule Number"] == rule, "Implemented"] = status

    new_order = ["Rule Number", "Section", "Description", "Implemented", "Type"]
    rules_data = rules_data.reindex(columns=new_order)
    rules_data = rules_data.sort_values("Rule Number")

    rules_data["Rule Number"] = rules_data["Rule Number"].apply(
        get_rule_path,
        benchmark_id=benchmark_id,
        implemented_rules=implemented_rules
    )

    # Convert DataFrame to Markdown table
    table = rules_data.to_markdown(index=False, tablefmt="github")

    # Add table title
    total_implemented = len([rule for rule, status in implemented_rules.items() if status == ":white_check_mark:"])
    status = total_implemented / len(implemented_rules)
    description = f"### {total_implemented}/{len(implemented_rules)} implemented rules ({status:.0%})\n\n"

    return table, description, status


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
            badge = f"[![CIS {service.upper()}]({badge_api}/CIS-Kubernetes%20({percentage:.0f}%25)-326CE5?" \
                    f"logo=Kubernetes)](RULES.md#k8s-cis-benchmark)\n"
        elif service == "eks":
            badge = f"[![CIS {service.upper()}]({badge_api}/CIS-Amazon%20EKS%20({percentage:.0f}%25)-FF9900?" \
                    f"logo=Amazon+EKS)](RULES.md#eks-cis-benchmark)\n"
        elif service == "aws":
            badge = f"[![CIS {service.upper()}]({badge_api}/CIS-AWS%20({percentage:.0f}%25)-232F3E?l" \
                    f"ogo=Amazon+AWS)](RULES.md#aws-cis-benchmark)\n"

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
    os.chdir(common.repo_root.working_dir + "/dev")

    # Write Markdown table to file
    with open("../RULES.md", "w") as f:
        f.write(f"# Rules Status")
        for benchmark_id in common.benchmark.keys():
            print(f"Generating Markdown table for '{benchmark_id}' service")
            f.write(f"\n\n## {benchmark_id.removeprefix('cis_').upper()} CIS Benchmark\n\n")
            table, description, percentage = generate_md_table(benchmark_id)
            f.write(description)
            f.write(table)
            update_main_readme_status_badge(percentage * 100, benchmark_id.removeprefix('cis_'))
        f.write("\n")
