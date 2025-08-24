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

def fix_broken_json_block(text: str) -> str:
    """
    Fixes a broken JSON block inside a Markdown triple backtick section.
    Ensures:
    - Balanced curly braces
    - Each closing brace is on its own line
    - Strips outer single quotes if needed

    Args:
        text (str): Markdown text containing the broken JSON block.

    Returns:
        str: Updated text with fixed JSON formatting.
    """
    pattern = re.compile(r"```(?:\s*)'?\{.*?```", re.DOTALL)
    match = pattern.search(text)

    if not match:
        return text

    broken_block = match.group(0)

    # Remove backticks and optional surrounding quotes
    json_content = broken_block.strip('`').strip()
    if json_content.startswith("'"):
        json_content = json_content[1:]
    if json_content.endswith("'"):
        json_content = json_content[:-1]

    # Count braces
    open_braces = json_content.count('{')
    close_braces = json_content.count('}')
    missing = open_braces - close_braces
    if missing > 0:
        json_content += '}' * missing

    # Split into lines and process closing braces to be on their own lines
    lines = json_content.splitlines()
    processed_lines = []

    for line in lines:
        line = line.strip()
        # If a line ends with multiple closing braces (like }}}), split them
        if re.match(r'^}+$', line):
            for char in line:
                processed_lines.append(char)
        else:
            # If a line ends with a closing brace (e.g., ...false"}}), split those braces
            brace_match = re.match(r'^(.*?)(}+)$', line)
            if brace_match:
                content, braces = brace_match.groups()
                if content.strip():
                    processed_lines.append(content.strip())
                for b in braces:
                    processed_lines.append('}')
            else:
                processed_lines.append(line)

    # Reassemble
    final_json = '\n'.join(processed_lines)
    final_block = f"```\n{final_json}\n```"
    return text[:match.start()] + final_block + text[match.end():]

def replacer(match):
    key = match.group(1)
    value = match.group(2)
    return f'{key}: "{value}"'

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

def fix_and_format_json(json_candidate):
    try:
        if json_candidate.startswith("'") and json_candidate.endswith("'"):
            json_candidate = json_candidate[1:-1]
        json_like_str = re.sub(r'(<[^>]+>)"', r'\1', json_candidate)

        json_like_str = re.sub(r'(".*?")\s*:\s*(<[^">]+>)', r'\1: "\2"', json_like_str)
        parsed = json.loads(json_like_str)
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

def format_and_fix_json_in_text(text):
    # Match code blocks that look like JSON
    text = fix_broken_json_block(text)
    pattern = r"```(?:json)?\s*({.*?})\s*```"
    matches = list(re.finditer(pattern, text, re.DOTALL))
    for match in reversed(matches):  # Reverse to avoid messing up indices
        original_block = match.group(0)
        json_candidate = match.group(1)
        formatted_json = fix_and_format_json(json_candidate)
        formatted_block = f"```json\n{formatted_json}\n```"
        text = text[:match.start()] + formatted_block + text[match.end():]

    return text

def format_json_in_text(text):
    try:
        # Match code blocks using triple backticks
        code_blocks = re.findall(r"```(?:json)?(.*?)```", text, re.DOTALL)
        for block in code_blocks:
            stripped_block = block.strip()
            try:
                parsed_json = json.loads(stripped_block)
                formatted_json = json.dumps(parsed_json, indent=4)
                # Replace original block with formatted block, tagged as JSON
                text = text.replace(f"```{block}```", f"```json\n{formatted_json}\n```")
            except json.JSONDecodeError:
                continue  # Not valid JSON, skip formatting
        return text
    except Exception:
        return text

def format_json_in_string_command(text):
    try:
        # Search for JSON-like content in the text
        start_index = text.find("{")
        end_index = text.rfind("}") + 1
        json_str = text[start_index:end_index]

        # Try to load and format the JSON
        parsed_json = json.loads(json_str)
        formatted_json = json.dumps(parsed_json, indent=4)

        # Replace the original JSON string in the text with the formatted one
        formatted_text = text[:start_index] + formatted_json + text[end_index:]

        return formatted_text
    except:
        # If JSON extraction or formatting fails, return the original text
        return text

def fix_code_blocks(text: str, rule_number: str, benchmark_id: str):
    text = add_new_line_after_period(text)
    if (rule_number in {"1.17", "3.5"} and benchmark_id == "cis_aws"):
        text = format_json_in_text(text)
    elif (rule_number in {"3.10", "3.11"} and benchmark_id == "cis_aws") or (rule_number in {"5.1.5"} and benchmark_id == "cis_azure"):
        text = format_json_in_string_command(text)
    else:
        text = format_and_fix_json_in_text(text)
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
