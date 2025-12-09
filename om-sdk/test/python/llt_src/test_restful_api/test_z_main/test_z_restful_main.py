import json

import pytest

from test_restful_api.test_z_main.restful_test_adapter import RestfulTestAdapter
from test_restful_api.test_z_main.restful_test_base import RestfulTestBase


class TestRestfulClient:
    # restful接口均设置了前缀,需要在url前加上
    test_adapter = RestfulTestAdapter()
    flask_client = test_adapter.get_flask_test_client()
    url_prefix = test_adapter.get_url_prefix()

    def setup_class(self):
        self.test_adapter.setup_class()

    def teardown_class(self):
        self.test_adapter.teardown_class()

    @pytest.mark.parametrize("test_case", RestfulTestBase.tests_list, ids=RestfulTestBase.labels_list)
    def test_restful_client(self, test_case: RestfulTestBase):
        url = test_case.get_url()
        method = test_case.get_method()
        data = test_case.get_data()
        header = test_case.get_header()
        expect_code = test_case.get_code()
        call_back_func = test_case.get_call_back()
        # 前置处理
        test_case.before()
        send_url = self.url_prefix + url
        test_response = self.flask_client.open(send_url, method=method, data=json.dumps(data), headers=header)
        ret_code = test_response.status_code
        assert ret_code == expect_code
        # 回调判断函数，需要在函数内部实现assert的判断
        call_back_func(test_response.get_data().decode())
        test_case.after()