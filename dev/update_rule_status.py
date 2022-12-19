import os
import pandas as pd
import git

"""
Generates Markdown tables with implemented rules status for all services.
"""

repo = git.Repo('.', search_parent_directories=True)
os.chdir(repo.working_dir + "/dev")

# List of compliance services
services = ["k8s", "eks", "aws"]

benchmark = {
    "k8s": "CIS_Kubernetes_V1.23_Benchmark_v1.0.1.xlsx",
    "eks": "CIS_Amazon_Elastic_Kubernetes_Service_(EKS)_Benchmark_v1.1.0.xlsx",
    "aws": "CIS_Amazon_Web_Services_Foundations_Benchmark_v1.5.0.xlsx",
}

relevant_sheets = {
    "k8s": ["Level 1 - Master Node", "Level 2 - Master Node", "Level 1 - Worker Node", "Level 2 - Worker Node"],
    "eks": ["MITRE & Controls Mappings"],
    "aws": ["MITRE ATT&CK Mappings"],
}

selected_columns_map = {
    "k8s": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title": "Description",
    },
    "eks": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title of Recommendation": "Description",
    },
    "aws": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title of Recommendation": "Description",
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
        excel_file = pd.read_excel(data_path, sheet_name=sheet_name, skiprows=1 if service == "aws" else 0)

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
    rules_status = get_implemented_rules(all_rules, service)

    # Add implemented rules' column to the data
    for rule, status in rules_status.items():
        full_data.loc[full_data["Rule Number"] == rule, "Implemented"] = status

    new_order = ["Rule Number", "Section", "Description", "Implemented"]
    full_data = full_data.reindex(columns=new_order)
    full_data = full_data.sort_values("Rule Number")

    # Convert DataFrame to Markdown table
    table = full_data.to_markdown(index=False, tablefmt="github")

    # Add table title
    total_implemented = len([rule for rule, status in rules_status.items() if status == ":white_check_mark:"])
    description = f"### {total_implemented}/{len(rules_status)} implemented rules  \n\n"

    return table, description


# Write Markdown table to file
with open("../RULES.md", "w") as f:
    f.write(f"# Rules Status")
    for service in services:
        print(f"Generating Markdown table for '{service}' service")
        f.write(f"\n\n## {service.upper()} CIS Benchmark\n\n")
        table, description = generate_md_table(service)
        f.write(description)
        f.write(table)
    f.write("\n")
