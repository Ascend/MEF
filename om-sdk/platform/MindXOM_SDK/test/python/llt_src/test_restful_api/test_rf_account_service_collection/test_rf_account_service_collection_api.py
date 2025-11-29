# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
from unittest.mock import patch

from test_restful_api.test_z_main.restful_test_base import GetTest

from user_manager.user_manager import UserManager


class TestGetSessionServiceCollection(GetTest):
    ACCOUNTS_SERVICE_COLLECTION_URL = "/redfish/v1/AccountService"

    def __init__(self, expect_ret, code: int, label: str):
        self.expect_ret = expect_ret
        self.patch = None
        super().__init__(url=self.ACCOUNTS_SERVICE_COLLECTION_URL,
                         code=code,
                         label=label,
                         )

    def before(self):
        self.patch = patch.object(UserManager, "get_all_info",
                                  return_value={"status": 200, "message": {"PasswordExpirationDays": 100}})
        self.patch.start()

    def after(self):
        if self.patch:
            self.patch.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def init_get_account_service_collection():
    TestGetSessionServiceCollection(
        expect_ret=json.dumps({
            "@odata.context": "/redfish/v1/$metadata#AccountService",
            "@odata.id": "/redfish/v1/AccountService",
            "@odata.type": "#AccountService.v1_11_0.AccountService",
            "Id": "AccountService", "Name": "Account Service", "PasswordExpirationDays": 100,
            "Accounts": {"@odata.id": "/redfish/v1/AccountService/Accounts"}
        }),
        code=200,
        label="Query password validity success.",
    ),


init_get_account_service_collection()