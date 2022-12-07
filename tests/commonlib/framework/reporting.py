"""
This module extends pytest basic report functionality using allure reporter
"""
from dataclasses import dataclass
import pytest
from allure_commons.types import LinkType
import allure


@dataclass
class SkipReportData:
    """
    SkipReportData class provides fields collection required for
    xfail, skip, and allure link markers
    """

    skip_reason: str = ""
    url_link: str = ""
    url_title: str = ""
    link_type: LinkType = LinkType.ISSUE


def skip_param_case(
    skip_list: list,
    data_to_report: SkipReportData,
    is_auto_defect: bool = True,
) -> list:
    """
    This function wraps parameterized test case(s) with markers:
    pytest.xfail
    allure.Link
    @param is_auto_defect:
    @param skip_list: list of test cases to be skipped
    @param data_to_report: Report data to be used in pytest and allure reports
    @return: list of test cases wrapped with xfail and allure link params.
    """
    ret_list = []
    if not skip_list:
        return ret_list

    marks_list = [
        pytest.mark.xfail(reason=data_to_report.skip_reason),
        allure.link(
            url=data_to_report.url_link,
            link_type=data_to_report.link_type,
            name=data_to_report.url_title,
        ),
    ]
    if is_auto_defect:
        marks_list.append(pytest.mark.skip(reason=data_to_report.skip_reason))

    for case in skip_list:
        ret_list.append(pytest.param(*case, marks=marks_list))
    return ret_list
