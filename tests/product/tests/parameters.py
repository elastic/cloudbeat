"""
Test suite parameters are registered here instead of being added directly
to tests, so they can be combined with cmdline arguments for dynamic parametrization.
"""


class Parameters:
    """
    Parameters of a test suite that can be used to generate test cases.
    """

    def __init__(self, argnames, argvalues, ids=None):
        self.argnames = argnames
        self.argvalues = argvalues
        self.ids = ids


TEST_PARAMETERS = {}


def register_params(func, params: Parameters):
    """
    Register test suite parameters for parametrization.
    """
    if func in TEST_PARAMETERS:
        raise KeyError(f"Parameters for test {func} are already registered")

    TEST_PARAMETERS[func] = params
