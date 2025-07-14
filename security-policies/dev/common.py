import json
import os

import git
import pandas as pd
import regex as re
from ruamel.yaml.scalarstring import PreservedScalarString as pss

repo_root = git.Repo(".", search_parent_directories=True)
rules_dir = os.path.join(
    repo_root.working_dir,
    "security-policies/bundle/compliance",
)

CODE_BLOCK_SIZE = 100

negative_emoji = ":x:"  # ❌
positive_emoji = ":white_check_mark:"  # ✅

benchmark = {
    "cis_k8s": "CIS_Kubernetes_V1.23_Benchmark_v1.0.1.xlsx",
    "cis_eks": "CIS_Amazon_Elastic_Kubernetes_Service_(EKS)_Benchmark_v1.0.1.xlsx",
    "cis_aws": "CIS_Amazon_Web_Services_Foundations_Benchmark_v1.5.0.xlsx",
    "cis_gcp": "CIS_Google_Cloud_Platform_Foundation_Benchmark_v2.0.0.xlsx",
    "cis_azure": "CIS_Microsoft_Azure_Foundations_Benchmark_v2.0.0.xlsx",
}

relevant_sheets = {
    "cis_k8s": [
        "Level 1 - Master Node",
        "Level 2 - Master Node",
        "Level 1 - Worker Node",
        "Level 2 - Worker Node",
    ],
    "cis_eks": ["Level 1", "Level 2"],
    "cis_aws": ["Level 1", "Level 2"],
    "cis_gcp": ["Level 1", "Level 2"],
    "cis_azure": ["Level 1", "Level 2"],
}

default_selected_columns_map = {
    "cis_k8s": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title": "Title",
        "Assessment Status": "Type",
    },
    "cis_eks": {
        "section #": "Section",
        "recommendation #": "Rule Number",
        "title": "Title",
        "scoring status": "Type",
    },
    "cis_aws": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title": "Title",
        "Assessment Status": "Type",
    },
    "cis_gcp": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title": "Title",
        "Assessment Status": "Type",
    },
    "cis_azure": {
        "Section #": "Section",
        "Recommendation #": "Rule Number",
        "Title": "Title",
        "Assessment Status": "Type",
    },
}


def status_emoji(positive):
    if positive:
        return positive_emoji
    return negative_emoji


def parse_rules_data_from_excel(
    benchmark_id,
    selected_columns=None,
    selected_rules=None,
):
    """
    Parse rules data from Excel file for current service.
    :param selected_rules: List of rules to parse
    :param selected_columns: Dictionary with columns to select from the sheet
    :param benchmark_id: Benchmark ID
    :return: Pandas DataFrame with rules data for current service and sections
    """
    if selected_columns is None:
        selected_columns = default_selected_columns_map

    benchmark_name = benchmark[benchmark_id]
    input_path = f"input/{benchmark_name}"

    sheets = relevant_sheets[benchmark_id]
    rules_data = pd.DataFrame()
    sections_df = pd.DataFrame()
    for sheet_name in sheets:
        print(f"Processing sheet '{sheet_name}'")
        excel_file = pd.read_excel(input_path, sheet_name=sheet_name)

        # Select only the columns you want to include in the Markdown table
        data = excel_file[selected_columns[benchmark_id].keys()]

        # Update Table headers
        data.columns = selected_columns[benchmark_id].values()

        # Remove rows with empty values in the "Rule Number" column and convert to string
        sections_curr_sheet = data.loc[
            data["Rule Number"].isna(),
            ["Section", "Title"],
        ].astype(str)

        # Filter out section information
        data = data[data["Rule Number"].notna()].astype(str)

        # Only keep the rules that are selected
        if selected_rules is not None:
            data = data[data["Rule Number"].isin(selected_rules)]

        # Add a new column with the sheet name
        data = data.assign(profile_applicability=sheet_name)

        rules_data = pd.concat([rules_data, data]).drop_duplicates(subset="Rule Number").reset_index(drop=True)
        sections_df = (
            pd.concat([sections_df, sections_curr_sheet])
            .drop_duplicates(subset="Section")
            .reset_index(
                drop=True,
            )
        )

    sections = {section: title for section, title in sections_df.values}

    return rules_data, sections


def check_and_fix_numbered_list(text):
    # Split the text into lines
    lines = text.split("\n")

    # Find the lines that start with a number and a period, and store their indices
    numbered_lines = [(i, line) for i, line in enumerate(lines) if re.match(r"^\d+\.", line)]

    # Check if the numbered lines are consecutively numbered
    for i, (index, line) in enumerate(numbered_lines):
        # Extract the number from the line
        line_number = int(line.split(".")[0])

        # Check if the line number is correct
        if line_number != i + 1:
            # The line number is not correct, fix it by replacing the line with the correct line number
            corrected_line = f"{i + 1}. {line.removeprefix(str(line_number) + '. ')}"
            lines[index] = corrected_line

    # Join the lines back into a single string and return the result
    return "\n".join(lines)


def add_new_line_after_period(text):
    # Split the text into lines
    lines = text.split("\n")

    # Find the lines that start with a number and a period
    numbered_lines = [line for line in lines if re.match(r"^\d+\.", line)]

    # Iterate through the lines and add a new line after a period, unless the line is a numbered line
    for i, line in enumerate(lines):
        if line not in numbered_lines:
            lines[i] = line.replace(". ", ".\n")

    # Join the lines back into a single string and return the result
    return "\n".join(lines)

def format_json_in_text(text):
    def fix_and_format_json(json_candidate):
        try:
            # Attempt to load directly
            parsed = json.loads(json_candidate)
        except json.JSONDecodeError:
            try:
                # Try to clean up invalid JSON-like text
                fixed = json_candidate.replace("'", '"')
                fixed = re.sub(r'(\w+):', r'"\1":', fixed)  # unquoted keys
                fixed = re.sub(r',\s*}', '}', fixed)         # trailing comma
                fixed = re.sub(r',\s*]', ']', fixed)         # trailing comma
                parsed = json.loads(fixed)
            except Exception:
                return json_candidate  # Return original if we can't fix
        return json.dumps(parsed, indent=4)

    # Match code blocks that look like JSON
    pattern = r"```(?:json)?\s*({.*?})\s*```"
    matches = list(re.finditer(pattern, text, re.DOTALL))

    for match in reversed(matches):  # Reverse to avoid messing up indices
        original_block = match.group(0)
        json_candidate = match.group(1)
        formatted_json = fix_and_format_json(json_candidate)
        formatted_block = f"```json\n{formatted_json}\n```"
        text = text[:match.start()] + formatted_block + text[match.end():]

    return text

def fix_code_blocks(text: str):
    text = add_new_line_after_period(text)
    text = format_json_in_text(text)
    return check_and_fix_numbered_list(text)


def apply_pss_recursively(data):
    if isinstance(data, dict):
        return {key: apply_pss_recursively(value) for key, value in data.items()}
    elif isinstance(data, list):
        return [value for value in data]
    elif isinstance(data, str):
        return pss(data) if len(data) > CODE_BLOCK_SIZE else data
    else:
        return data

