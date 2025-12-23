from functools import wraps


class FakeFunc:
    def __call__(self, *args, **kwargs):
        return 0


class FakeCDLL:
    def __getattr__(self, item):
        func = FakeFunc()
        setattr(self, item, func)
        return func


class MockPrivilegeAuth:
    @staticmethod
    def token_required(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper
