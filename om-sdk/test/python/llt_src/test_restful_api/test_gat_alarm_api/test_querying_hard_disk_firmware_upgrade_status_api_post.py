import json

from test_restful_api.test_z_main.restful_test_base import PostTest


class TestCreateNewSessionFailed(PostTest):
    """2.4.4 查询硬盘固件升级状态信息"""
    COMPUTER_SYSTEM_RESET_URL = "/redfish/v1/UpdateHddService/Actions/UpdateHddService.infos"

    def __init__(self, expect_ret, code: int, label: str, data: dict):
        self.expect_ret = expect_ret
        self.data = data
        super().__init__(url=self.COMPUTER_SYSTEM_RESET_URL,
                         code=code,
                         data=data,
                         label=label)

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def test_querying_hard_disk_firmware_upgrade_status():
    TestCreateNewSessionFailed(
        expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": None}}]}}),
        label="test querying hard_disk firmware upgrade status failed \
            due to Parameter is invalid for HddNo must be 1 or 2.",
        code=400,
        data={
            "HddNo": 3
        })
