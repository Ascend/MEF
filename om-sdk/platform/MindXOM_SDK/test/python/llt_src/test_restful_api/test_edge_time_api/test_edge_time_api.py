# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json

from test_restful_api.test_z_main.restful_test_base import GetTest


class TestGetEdgeTime(GetTest):
    EDGE_TIME_URL = "/redfish/v1/Systems/SystemTime"

    def __init__(self, expect_ret, code: int, label: str = "Test Get System Time"):
        super().__init__(url=self.EDGE_TIME_URL, code=code, label=label)
        self.expect_ret = expect_ret

    def call_back_assert(self, test_response: str):
        # 该接口为获取实时时间接口，无法准确到时间匹配，获取成功Datetime里面会有值
        dict_response = json.loads(test_response)
        datetime = dict_response.get(self.expect_ret)
        # 判断接收到的参数里面有值
        assert datetime


def init_get_edge_time_instances():
    TestGetEdgeTime("Datetime", code=200)


init_get_edge_time_instances()