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
CIS_4_3_2 = "CIS 4.3.2"
CIS_4_3_3 = "CIS 4.3.3"
CIS_4_3_4 = "CIS 4.3.4"
CIS_4_3_5 = "CIS 4.3.5"
CIS_4_3_6 = "CIS 4.3.6"
CIS_4_3_7 = "CIS 4.3.7"
CIS_4_3_8 = "CIS 4.3.8"
# Disable 4.4.1 - Azure Database for MySQL - Single Server is being retired
# See: https://learn.microsoft.com/en-us/azure/mysql/single-server/whats-happening-to-mysql-single-server
# CIS_4_4_1 = "CIS 4.4.1"
CIS_4_4_2 = "CIS 4.4.2"
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

cis_azure_4_3_2_pass = AzureServiceCase(
    rule_tag=CIS_4_3_2,
    case_identifier="test-pgdb-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_3_2_fail = AzureServiceCase(
    rule_tag=CIS_4_3_2,
    case_identifier="test-pgdb-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_3_2 = {
    """4.3.2 Ensure Server Parameter 'log_checkpoints' is set to 'ON' for
        PostgreSQL Database Server (Automated) expect: passed""": cis_azure_4_3_2_pass,
    """4.3.2 Ensure Server Parameter 'log_checkpoints' is set to 'ON' for
        PostgreSQL Database Server (Automated) expect: failed""": cis_azure_4_3_2_fail,
}

cis_azure_4_3_3_pass = AzureServiceCase(
    rule_tag=CIS_4_3_3,
    case_identifier="test-pgdb-pass",
    expected=RULE_PASS_STATUS,
)

# TODO: This will be cleaned up in issue https://github.com/elastic/cloudbeat/issues/2544
# cis_azure_4_3_3_fail = AzureServiceCase(
#     rule_tag=CIS_4_3_3,
#     case_identifier="test-postgresql-single-server-failpgserver",
#     expected=RULE_FAIL_STATUS,
# )

cis_azure_4_3_3 = {
    """4.3.3 Ensure server parameter 'log_connections' is set to 'ON' for
        PostgreSQL Database Server (Automated) expect: passed""": cis_azure_4_3_3_pass,
    # TODO: This will be cleaned up in issue https://github.com/elastic/cloudbeat/issues/2544
    # """4.3.3 Ensure server parameter 'log_connections' is set to 'ON' for
    #     PostgreSQL Database Server (Automated) expect: failed""": cis_azure_4_3_3_fail,
}

cis_azure_4_3_4_pass = AzureServiceCase(
    rule_tag=CIS_4_3_4,
    case_identifier="test-pgdb-pass",
    expected=RULE_PASS_STATUS,
)

# TODO: This will be cleaned up in issue https://github.com/elastic/cloudbeat/issues/2544
# cis_azure_4_3_4_fail = AzureServiceCase(
#     rule_tag=CIS_4_3_4,
#     case_identifier="test-postgresql-single-server-failpgserver",
#     expected=RULE_FAIL_STATUS,
# )

cis_azure_4_3_4 = {
    """4.3.4 Ensure server parameter 'log_disconnections' is set to 'ON' for
        PostgreSQL Database Server (Automated) expect: passed""": cis_azure_4_3_4_pass,
    # TODO: This will be cleaned up in issue https://github.com/elastic/cloudbeat/issues/2544
    # """4.3.4 Ensure server parameter 'log_disconnections' is set to 'ON' for
    #     PostgreSQL Database Server (Automated) expect: failed""": cis_azure_4_3_4_fail,
}

# TODO: This will be cleaned up in issue https://github.com/elastic/cloudbeat/issues/2544
# cis_azure_4_3_5_pass_single_server = AzureServiceCase(
#     rule_tag=CIS_4_3_5,
#     case_identifier="test-postgresql-single-server",
#     expected=RULE_PASS_STATUS,
# )

cis_azure_4_3_5_fail_single_server = AzureServiceCase(
    rule_tag=CIS_4_3_5,
    case_identifier="test-postgresql-single-server-failpgserver",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_3_5_pass_flexible_server = AzureServiceCase(
    rule_tag=CIS_4_3_5,
    case_identifier="test-pgdb-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_3_5_fail_flexible_server = AzureServiceCase(
    rule_tag=CIS_4_3_5,
    case_identifier="test-pgdb-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_4_3_5 = {
    # TODO: This will be cleaned up in issue https://github.com/elastic/cloudbeat/issues/2544
    # """4.3.5 Ensure server parameter 'connection_throttling' is set to 'ON' for PostgreSQL Database Server
    # (Automated) [SINGLE SERVER] expect: passed""": cis_azure_4_3_5_pass_single_server,
    """4.3.5 Ensure server parameter 'connection_throttling' is set to 'ON' for PostgreSQL Database Server
    (Automated) [SINGLE SERVER] expect: failed""": cis_azure_4_3_5_fail_single_server,
    """4.3.5 Ensure server parameter 'connection_throttling' is set to 'ON' for PostgreSQL Database Server
    (Automated) [FLEXIBLE SERVER] expect: passed""": cis_azure_4_3_5_pass_flexible_server,
    """4.3.5 Ensure server parameter 'connection_throttling' is set to 'ON' for PostgreSQL Database Server
    (Automated) [FLEXIBLE SERVER] expect: failed""": cis_azure_4_3_5_fail_flexible_server,
}

# TODO: This will be cleaned up in issue https://github.com/elastic/cloudbeat/issues/2544
# cis_azure_4_3_6_pass = AzureServiceCase(
#     rule_tag=CIS_4_3_6,
#     case_identifier="test-postgresql-single-server",
#     expected=RULE_PASS_STATUS,
# )

# cis_azure_4_3_6_fail = AzureServiceCase(
#     rule_tag=CIS_4_3_6,
#     case_identifier="test-postgresql-single-server-failpgserver",
#     expected=RULE_FAIL_STATUS,
# )

# cis_azure_4_3_6 = {
#     """4.3.6 Ensure Server Parameter 'log_retention_days' is greater
#         than 3 days for PostgreSQL Database Server (Automated) expect: passed""": cis_azure_4_3_6_pass,
#     """4.3.6 Ensure Server Parameter 'log_retention_days' is greater
#         than 3 days for PostgreSQL Database Server (Automated) expect: failed""": cis_azure_4_3_6_fail,
# }

cis_azure_4_3_7_pass = AzureServiceCase(
    rule_tag=CIS_4_3_7,
    case_identifier="test-pgdb-pass",
    expected=RULE_PASS_STATUS,
)

# TODO: This will be cleaned up in issue https://github.com/elastic/cloudbeat/issues/2544
# cis_azure_4_3_7_fail = AzureServiceCase(
#     rule_tag=CIS_4_3_7,
#     case_identifier="test-postgresql-single-server-failpgserver",
#     expected=RULE_FAIL_STATUS,
# )

cis_azure_4_3_7 = {
    """4.3.7 Ensure 'Allow access to Azure services' for PostgreSQL
        Database Server is disabled (Automated) expect: passed""": cis_azure_4_3_7_pass,
    # TODO: This will be cleaned up in issue https://github.com/elastic/cloudbeat/issues/2544
    # """4.3.7 Ensure 'Allow access to Azure services' for PostgreSQL
    #     Database Server is disabled (Automated) expect: failed""": cis_azure_4_3_7_fail,
}

# TODO: This will be cleaned up in issue https://github.com/elastic/cloudbeat/issues/2544
# cis_azure_4_3_8_fail = AzureServiceCase(
#     rule_tag=CIS_4_3_8,
#     case_identifier="test-postgresql-single-server-failpgserver",
#     expected=RULE_FAIL_STATUS,
# )

# cis_azure_4_3_8 = {
#     # Can't test this rule passing, motivation: https://github.com/elastic/cloudbeat/pull/1797
#     """4.3.8 Ensure 'Infrastructure double encryption' for PostgreSQL
#         Database Server is 'Enabled' (Automated) expect: failed""": cis_azure_4_3_8_fail,
# }

# 4.4.* Rules ====================================

# cis_azure_4_4_1_pass = AzureServiceCase(
#     rule_tag=CIS_4_4_1,
#     case_identifier="rule-441",
#     expected=RULE_PASS_STATUS,
# )
#
# cis_azure_4_4_1_fail = AzureServiceCase(
#     rule_tag=CIS_4_4_1,
#     case_identifier="rule-441-fail",
#     expected=RULE_FAIL_STATUS,
# )
#
# cis_azure_4_4_1 = {
#     """4.4.1 Ensure 'Enforce SSL connection' is set to 'Enabled'
#       for Standard MySQL Database Server expect: passed""": cis_azure_4_4_1_pass,
#     """4.4.1 Ensure 'Enforce SSL connection' is set to 'Enabled'
#       for Standard MySQL Database Server expect: failed""": cis_azure_4_4_1_fail,
# }

cis_azure_4_4_2_pass = AzureServiceCase(
    rule_tag=CIS_4_4_2,
    case_identifier="test-mysql-db-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_4_4_2 = {
    # Can't test this rule failing, motivation: https://github.com/elastic/cloudbeat/pull/1811
    """4.4.2 Ensure 'TLS Version' is set to 'TLSV1.2' for MySQL flexible
        Database Server (Automated) expect: passed""": cis_azure_4_4_2_pass,
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
    **cis_azure_4_3_2,
    **cis_azure_4_3_3,
    **cis_azure_4_3_4,
    **cis_azure_4_3_5,
    # **cis_azure_4_3_6,
    **cis_azure_4_3_7,
    # **cis_azure_4_3_8,
    # **cis_azure_4_4_1,
    **cis_azure_4_4_2,
    **cis_azure_4_5_1,
}
