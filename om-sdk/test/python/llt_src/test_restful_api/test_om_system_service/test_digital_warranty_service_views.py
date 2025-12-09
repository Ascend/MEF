import json

from mock import patch

from test_restful_api.test_z_main.restful_test_base import GetTest


class TestGetDigital(GetTest):
    DIGITAL_URL = "/redfish/v1/Systems/DigitalWarranty"

    def __init__(self, expect_ret, code: int, label: str):
        self.expect_ret = expect_ret
        self.patch = None
        super().__init__(url=self.DIGITAL_URL, code=code, label=label)

    def before(self):
        self.patch = patch("lib_restful_adapter.LibRESTfulAdapter.lib_restful_interface",
                           return_value={"status": 200, "message": {}})
        self.patch.start()

    def after(self):
        if self.patch:
            self.patch.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret in test_response


def init_test_get_digital():
    TestGetDigital(expect_ret="ItemId",
                   label="test get digital success!",
                   code=200,
                   )


init_test_get_digital()
