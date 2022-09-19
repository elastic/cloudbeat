class Parameters:
    def __init__(self, argnames, argvalues, ids=None):
        self.argnames = argnames
        self.argvalues = argvalues
        self.ids = ids


TEST_PARAMETERS = {}


def register_params(func, params: Parameters):
    if func in TEST_PARAMETERS:
        raise KeyError(f'Parameters for test {func} are already registered')

    TEST_PARAMETERS[func] = params
