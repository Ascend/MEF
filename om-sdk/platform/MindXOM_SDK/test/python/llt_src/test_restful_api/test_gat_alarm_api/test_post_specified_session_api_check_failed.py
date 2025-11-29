# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json

from test_restful_api.test_z_main.restful_test_base import PostTest


class TestPostSessionFailed(PostTest):
    ACCOUNTS_URL = "/redfish/v1/SessionService/Sessions"

    def __init__(self, expect_ret, data, code: int, label: str):
        self.expect_ret = expect_ret
        super().__init__(url=self.ACCOUNTS_URL,
                         code=code,
                         label=label,
                         data=data)

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def init_post_session_failed():
    TestPostSessionFailed(
        expect_ret=json.dumps(
            {"error": {
                "code": "Base.1.0.MalformedJSON",
                "message": "A MalformedJSON has occurred. See ExtendedInfo for more information.",
                "@Message.ExtendedInfo": [
                    {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                     "Description": "Indicates that the request body was malformed JSON.  "
                                    "Could be duplicate, syntax error,etc.",
                     "Message": "The request body submitted was malformed JSON "
                                "and could not be parsed by the receiving service.",
                     "Severity": "Critical", "NumberOfArgs": 0, "ParamTypes": None,
                     "Resolution": "Ensure that the request body is valid JSON and resubmit the request.",
                     "Oem": {"status": None}}]}}
        ),
        code=400,
        data={},
        label="test delete specified session failed due to data is None."),

    TestPostSessionFailed(
        expect_ret=json.dumps(
            {"error": {
                "code": "Base.1.0.MalformedJSON",
                "message": "A MalformedJSON has occurred. See ExtendedInfo for more information.",
                "@Message.ExtendedInfo": [
                    {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                     "Description": "Indicates that the request body was malformed JSON.  "
                                    "Could be duplicate, syntax error,etc.",
                     "Message": "The request body submitted was malformed JSON "
                                "and could not be parsed by the receiving service.",
                     "Severity": "Critical", "NumberOfArgs": 0, "ParamTypes": None,
                     "Resolution": "Ensure that the request body is valid JSON and resubmit the request.",
                     "Oem": {"status": None}}]}}
        ),
        code=400,
        data="test",
        label="test delete specified session failed due to data is not json."),

    TestPostSessionFailed(
        expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The user name or password error.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 110207}}]}}),
        code=400,
        data={"UserName": "", "Password": "test"},
        label="test delete specified session failed due to UserName is null."),


init_post_session_failed()
