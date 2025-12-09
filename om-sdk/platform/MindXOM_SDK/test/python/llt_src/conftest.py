from _pytest.python import Metafunc


def pytest_generate_tests(metafunc: Metafunc):
    """
    pytest 预处理用例hook函数。启动测试时pytest自动加载。
    """
    model = metafunc.function.__annotations__.get("model")
    if not model:
        return
    use_cases = metafunc.cls.use_cases[metafunc.function.__name__]
    if not use_cases:
        return
    models = []
    names = []
    for name, value in use_cases.items():
        names.append(name)
        models.append(model(*value) if not isinstance(value, model) else value)
    metafunc.parametrize("model", models, ids=names)


class TestBase:
    # 测试用例数据，写好后会由pytest_generate_tests自动加载.key为用例名，值为用例数据。
    use_cases: dict
