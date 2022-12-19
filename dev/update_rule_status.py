import os
import pandas as pd
import git

"""
Generates Markdown tables with implemented rules status for all services.
"""

repo = git.Repo('.', search_parent_directories=True)
os.chdir(repo.working_dir + "/dev")

benchmark = {
    "k8s": "CIS_Kubernetes_V1.23_Benchmark_v1.0.1.xlsx",
    "eks": "CIS_Amazon_Elastic_Kubernetes_Service_(EKS)_Benchmark_v1.1.0.xlsx",
    "aws": "CIS_Amazon_Web_Services_Foundations_Benchmark_v1.5.0.xlsx",
}

relevant_sheets = {
    "k8s": ["Level 1 - Master Node", "Level 2 - Master Node", "Level 1 - Worker Node", "Level 2 - Worker Node"],
    "eks": ["Level 1", "Level 2"],
    "aws": ["Level 1", "Level 2"],
}

selected_columns_map = {
    "k8s": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title": "Description",
        "Assessment Status": "Type",
    },
    "eks": {
        "section #": "Section",
        "recommendation #": "Rule Number",
        "title": "Description",
        "assessment status": "Type",
    },
    "aws": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title": "Description",
        "Assessment Status": "Type",
    },
}


def get_implemented_rules(all_rules, service):
    """
    Get list of implemented rules in the repository for current service.
    :param all_rules: List of all rules for specified benchmark
    :param service: Service name (k8s, eks, aws)
    :return: List of implemented rules
    """
    # Set all rules as not implemented by default
    implemented_rules = {str(rule): ":x:" for rule in all_rules}  # ❌

    # Construct path to rules directory for current service
    rules_dir = os.path.join("../bundle", "compliance", f"cis_{service}", "rules")

    # Get list of all rule files in the rules directory
    rule_files = os.listdir(rules_dir)

    # Iterate over all rule files
    for rule_file in rule_files:
        # Extract rule number from rule file name
        rule_number = rule_file.removeprefix("cis_").replace("_", ".")

        # Set rule as implemented
        implemented_rules[rule_number] = ":white_check_mark:"  # ✅

    return implemented_rules


def generate_md_table(service):
    """
    Generate Markdown table with implemented rules status for current service.
    :param service: Service name (k8s, eks, aws)
    :return: Markdown table
    """
    benchmark_name = benchmark[service]
    data_path = f"../cis_policies_generator/input/{benchmark_name}"

    sheets = relevant_sheets[service]
    full_data = pd.DataFrame()
    for sheet_name in sheets:
        print(f"Processing sheet '{sheet_name}'")
        excel_file = pd.read_excel(data_path, sheet_name=sheet_name)

        # Select only the columns you want to include in the Markdown table
        data = excel_file[selected_columns_map[service].keys()]

        # Update Table headers
        data.columns = selected_columns_map[service].values()

        # Remove rows with empty values in the "Rule Number" column and convert to string
        data = data[data["Rule Number"].notna()].astype(str)

        full_data = pd.concat([full_data, data]).drop_duplicates(subset="Rule Number").reset_index(drop=True)

    # Get list of all rules in sheet
    all_rules = full_data["Rule Number"].to_list()

    # Get list of implemented rules
    implemented_rules = get_implemented_rules(all_rules, service)

    # Add implemented rules' column to the data
    for rule, status in implemented_rules.items():
        full_data.loc[full_data["Rule Number"] == rule, "Implemented"] = status

    new_order = ["Rule Number", "Section", "Description", "Implemented", "Type"]
    full_data = full_data.reindex(columns=new_order)
    full_data = full_data.sort_values("Rule Number")

    full_data["Rule Number"] = full_data["Rule Number"].apply(get_rule_path, service=service, implemented_rules=implemented_rules)

    # Convert DataFrame to Markdown table
    table = full_data.to_markdown(index=False, tablefmt="github")

    # Add table title
    total_implemented = len([rule for rule, status in implemented_rules.items() if status == ":white_check_mark:"])
    status = total_implemented / len(implemented_rules)
    description = f"### {total_implemented}/{len(implemented_rules)} implemented rules ({status:.0%})\n\n"

    return table, description, status


def get_rule_path(rule, service, implemented_rules):
    """
    Get rule path for specified rule and service.
    :param implemented_rules: ‘Implemented’ column values
    :param rule: Rule number
    :param service: Service name (k8s, eks, aws)
    :return: Rule path in the repository
    """
    if implemented_rules[rule] == ":white_check_mark:":
        return f"[{rule}](bundle/compliance/cis_{service}/rules/cis_{rule.replace('.', '_')})"
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
            badge = f"[![CIS {service.upper()}]({badge_api}/CIS-Kubernetes%20({percentage:.1f}%25)-326CE5?" \
                    f"logo=Kubernetes)](RULES.md#k8s-cis-benchmark)\n"
        elif service == "eks":
            badge = f"[![CIS {service.upper()}]({badge_api}/CIS-Amazon%20EKS%20({percentage:.1f}%25)-FF9900?" \
                    f"logo=Amazon+EKS)](RULES.md#eks-cis-benchmark)\n"
        elif service == "aws":
            badge = f"[![CIS {service.upper()}]({badge_api}/CIS-AWS%20({percentage:.1f}%25)-232F3E?l" \
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


with open("../RULES.md", "w") as f:
    f.write(f"# Rules Status")
    for service in benchmark.keys():
        print(f"Generating Markdown table for '{service}' service")
        f.write(f"\n\n## {service.upper()} CIS Benchmark\n\n")
        table, description, percentage = generate_md_table(service)
        f.write(description)
        f.write(table)
        update_main_readme_status_badge(percentage * 100, service)
    f.write("\n")
