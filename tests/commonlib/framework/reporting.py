"""
This module extends pytest basic report functionality using allure reporter
"""

from __future__ import annotations

from dataclasses import dataclass

import allure
import pytest
from allure_commons.types import LinkType


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
    skip_objects: dict | list,
    data_to_report: SkipReportData,
    is_auto_defect: bool = True,
) -> dict | list | None:
    """
    This function wraps parameterized test case(s) with markers:
    pytest.xfail
    allure.Link
    @param is_auto_defect:
    @param skip_objects: dictionary or list of test cases to be skipped
    @param data_to_report: Report data to be used in pytest and allure reports
    @return: dictionary or list of test cases wrapped with xfail and allure link params.
    """

    if not skip_objects:
        return None

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

    if isinstance(skip_objects, list):
        ret_object = []
        for case in skip_objects:
            ret_object.append(pytest.param(*case, marks=marks_list))
    elif isinstance(skip_objects, dict):
        ret_object = {}
        for key, value in skip_objects.items():
            ret_object[key] = pytest.param(*value, marks=marks_list)
    else:
        ret_object = None

    return ret_object
