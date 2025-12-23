from abc import ABC

from test_restful_api.test_z_main.restful_test_base import RestfulTestBase


class RequestTest(RestfulTestBase, ABC):
    def __init__(self, request_obj: dict, response_obj: dict, label: str):
        super().__init__(
            url=request_obj.get("url") + request_obj.get("params"),
            code=response_obj.get("code"),
            data=request_obj.get("data"),
            label=label,
            method=request_obj.get("method"),
        )

        self.expect_response = response_obj.get("expect_response")

    def call_back_assert(self, test_response: str):
        assert self.expect_response == test_response
