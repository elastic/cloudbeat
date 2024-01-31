"""
This module provides Azure database service rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Database service identification is performed by resource name.
"""

from ..azure_test_case import AzureServiceCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

CIS_4_1_1 = "CIS 4.1.1"
CIS_4_1_2 = "CIS 4.1.2"
CIS_4_1_3 = "CIS 4.1.3"
CIS_4_1_4 = "CIS 4.1.4"
CIS_4_1_5 = "CIS 4.1.5"
CIS_4_1_6 = "CIS 4.1.6"
CIS_4_2_1 = "CIS 4.2.1"
CIS_4_3_1 = "CIS 4.3.1"
CIS_4_4_1 = "CIS 4.4.1"
CIS_4_5_1 = "CIS 4.5.1"

# 4.1.* Rules ====================================

cis_azure_4_1_1_pass = AzureServiceCase(
    rule_tag=CIS_4_1_1,
    case_identifier="test-sql-db-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_1_1_fail = AzureServiceCase(
    rule_tag=CIS_4_1_1,
    case_identifier="test-sql-db-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_1_1 = {
    """4.1.1 Ensure that 'Auditing' is set to 'On' (Automated) expect: passed""": cis_azure_4_1_1_pass,
    """4.1.1 Ensure that 'Auditing' is set to 'On' (Automated) expect: failed""": cis_azure_4_1_1_fail,
}

cis_azure_4_1_2_pass = AzureServiceCase(
    rule_tag=CIS_4_1_2,
    case_identifier="test-sql-db-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_1_2_fail = AzureServiceCase(
    rule_tag=CIS_4_1_2,
    case_identifier="test-sql-db-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_1_2 = {
    """4.1.2 Ensure no Azure SQL Databases allow ingress
      from 0.0.0.0/0 (ANY IP) expect: passed""": cis_azure_4_1_2_pass,
    """4.1.2 Ensure no Azure SQL Databases allow ingress
      from 0.0.0.0/0 (ANY IP) expect: failed""": cis_azure_4_1_2_fail,
}

cis_azure_4_1_3_pass = AzureServiceCase(
    rule_tag=CIS_4_1_3,
    case_identifier="test-sql-db-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_1_3_fail = AzureServiceCase(
    rule_tag=CIS_4_1_3,
    case_identifier="test-sql-db-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_1_3 = {
    """4.1.3 Ensure SQL server's Transparent Data Encryption (TDE)
        protector is encrypted with Customer-managed key (Automated) expect: passed""": cis_azure_4_1_3_pass,
    """4.1.3 Ensure SQL server's Transparent Data Encryption (TDE)
        protector is encrypted with Customer-managed key (Automated) expect: fail""": cis_azure_4_1_3_fail,
}

cis_azure_4_1_4_pass = AzureServiceCase(
    rule_tag=CIS_4_1_4,
    case_identifier="test-sql-db-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_1_4_fail = AzureServiceCase(
    rule_tag=CIS_4_1_4,
    case_identifier="test-sql-db-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_1_4 = {
    """4.1.4 Ensure that Azure Active Directory Admin
      is Configured for SQL Servers expect: passed""": cis_azure_4_1_4_pass,
    """4.1.4 Ensure that Azure Active Directory Admin
      is Configured for SQL Servers expect: failed""": cis_azure_4_1_4_fail,
}

cis_azure_4_1_5_pass = AzureServiceCase(
    rule_tag=CIS_4_1_5,
    case_identifier="test-sql-db-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_1_5_fail = AzureServiceCase(
    rule_tag=CIS_4_1_5,
    case_identifier="test-sql-db-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_1_5 = {
    """4.1.5 Ensure that 'Data encryption' is set to 'On' on a SQL
        Database (Automated) expect: passed""": cis_azure_4_1_5_pass,
    """4.1.5 Ensure that 'Data encryption' is set to 'On' on a SQL
        Database (Automated) expect: failed""": cis_azure_4_1_5_fail,
}

cis_azure_4_1_6_pass = AzureServiceCase(
    rule_tag=CIS_4_1_6,
    case_identifier="test-sql-db-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_1_6_fail = AzureServiceCase(
    rule_tag=CIS_4_1_6,
    case_identifier="test-sql-db-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_1_6 = {
    """4.1.6 Ensure that 'Auditing' Retention is 'greater than 90 days'
        (Automated) expect: passed""": cis_azure_4_1_6_pass,
    """4.1.6 Ensure that 'Auditing' Retention is 'greater than 90 days'
        (Automated) expect: failed""": cis_azure_4_1_6_fail,
}

# 4.2.* Rules ====================================

cis_azure_4_2_1_pass = AzureServiceCase(
    rule_tag=CIS_4_2_1,
    case_identifier="test-sql-db-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_2_1_fail = AzureServiceCase(
    rule_tag=CIS_4_2_1,
    case_identifier="test-sql-db-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_2_1 = {
    """4.2.1 Ensure that Microsoft Defender for SQL is set to 'On' for
        critical SQL Servers (Automated) expect: passed""": cis_azure_4_2_1_pass,
    """4.2.1 Ensure that Microsoft Defender for SQL is set to 'On' for
        critical SQL Servers (Automated) expect: failed""": cis_azure_4_2_1_fail,
}

# 4.3.* Rules ====================================

cis_azure_4_3_1_pass = AzureServiceCase(
    rule_tag=CIS_4_3_1,
    case_identifier="test-postgresql-single-server",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_3_1_fail = AzureServiceCase(
    rule_tag=CIS_4_3_1,
    case_identifier="test-postgresql-single-server-failpgserver",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_3_1 = {
    """4.3.1 Ensure 'Enforce SSL connection' is set to 'ENABLED'
      for PostgreSQL Database Server expect: passed""": cis_azure_4_3_1_pass,
    """4.3.1 Ensure 'Enforce SSL connection' is set to 'ENABLED'
      for PostgreSQL Database Server expect: failed""": cis_azure_4_3_1_fail,
}

# 4.4.* Rules ====================================

cis_azure_4_4_1_pass = AzureServiceCase(
    rule_tag=CIS_4_4_1,
    case_identifier="rule-441",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_4_1_fail = AzureServiceCase(
    rule_tag=CIS_4_4_1,
    case_identifier="rule-441-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_4_1 = {
    """4.4.1 Ensure 'Enforce SSL connection' is set to 'Enabled'
      for Standard MySQL Database Server expect: passed""": cis_azure_4_4_1_pass,
    """4.4.1 Ensure 'Enforce SSL connection' is set to 'Enabled'
      for Standard MySQL Database Server expect: failed""": cis_azure_4_4_1_fail,
}

# 4.5.* Rules ====================================

cis_azure_4_5_1_pass = AzureServiceCase(
    rule_tag=CIS_4_5_1,
    case_identifier="test-cosmos-db-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_5_1_fail = AzureServiceCase(
    rule_tag=CIS_4_5_1,
    case_identifier="test-cosmos-db-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_5_1 = {
    """4.5.1 Ensure That 'Firewalls & Networks' Is Limited to
      Use Selected Networks Instead of All Networks expect: passed""": cis_azure_4_5_1_pass,
    """4.5.1 Ensure That 'Firewalls & Networks' Is Limited to
      Use Selected Networks Instead of All Networks expect: failed""": cis_azure_4_5_1_fail,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_azure_4_1_1,
    **cis_azure_4_1_2,
    **cis_azure_4_1_3,
    **cis_azure_4_1_4,
    **cis_azure_4_1_5,
    **cis_azure_4_1_6,
    **cis_azure_4_2_1,
    **cis_azure_4_3_1,
    **cis_azure_4_4_1,
    **cis_azure_4_5_1,
}
