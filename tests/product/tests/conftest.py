"""
This module provides fixtures and configurations for
product tests.
"""

from product.tests.parameters import TEST_PARAMETERS


def pytest_generate_tests(metafunc):
    """
    This function generates the test cases to run using the set of
    test cases registered in TEST_PARAMETERS and the values passed to
    relevant custom cmdline parameters such as --range.
    """
    # -k command line option to specify an expression which implements a substring match on the test names
    # instead of the exact match on markers that -m provides. This makes it easy to select tests based on their names
    if "-k" in metafunc.config.invocation_params.args:
        parametrize_eks_params(metafunc)
        return

    if (
        metafunc.definition.get_closest_marker(
            metafunc.config.getoption("markexpr", default=None),
        )
        is None
    ):
        return

    params = TEST_PARAMETERS.get(metafunc.function)
    if params is None:
        raise ValueError(f"Params for function {metafunc.function} are not registered.")

    test_range = metafunc.config.getoption("range")
    test_range_start, test_range_end = test_range.split("..")

    if test_range_end != "" and int(test_range_end) < len(params.argvalues):
        params.argvalues = params.argvalues[: int(test_range_end)]

        if params.ids is not None:
            params.ids = params.ids[: int(test_range_end)]

    if test_range_start != "":
        if int(test_range_start) >= len(params.argvalues):
            raise ValueError(f"Invalid range for test function {metafunc.function}")

        params.argvalues = params.argvalues[int(test_range_start) :]

        if params.ids is not None:
            params.ids = params.ids[int(test_range_start) :]

    metafunc.parametrize(params.argnames, params.argvalues, ids=params.ids)


def parametrize_eks_params(func_details):
    """
    This function creates parametrization for EKS test cases
    @param func_details: metafunc
    @return:
    """
    params = TEST_PARAMETERS.get(func_details.function)
    if params is None:
        raise ValueError(f"Params for function {func_details.function} are not registered.")

    func_details.parametrize(params.argnames, params.argvalues, ids=params.ids)
