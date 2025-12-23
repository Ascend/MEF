import ipaddress
from collections import namedtuple

import pytest
from pytest_mock import MockerFixture

from common.constants.base_constants import CommonConstants
from common.constants.error_codes import CommonErrorCodes
from common.exception.biz_exception import BizException
from common.utils.app_common_method import AppCommonMethod
from common.utils.exec_cmd import ExecCmd
from ut_utils.mock_utils import mock_npu_smi_board_type
from common.checkers import IPV4Checker, CheckResult

GetEthIpMap = namedtuple("GetEthIpMap", "expect, cmd")
GetEthGateway = namedtuple("GetEthGateway", "expect, ip, cmd, gateway, network")
GetGateway = namedtuple("GetGateway", "expect, cmd")
GetAllInfo = namedtuple("GetAllInfo", "expect, map")
PostRequest = namedtuple("PostRequest", "expect, lock, check_root")

with mock_npu_smi_board_type():
    from lib.Linux.systems.actions import RestoreDefaultsAction


class TestRestoreDefaultsActions:
    use_cases = {
        "test_get_eth_gateway": {
            "exception": (None, ["ip"], [1, ""], None, None),
            "null gateway": ([], ["ip"], [0, "/abc"], [], None),
            "normal": ([], ["ip"], [0, "/abc"], ["abc"], "address")
        },
        "test_post_request": {
            "busy": (
                [400, [CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED.code,
                       CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED.messageKey]], True, None
            ),
            "Restore default failed.": (
                [CommonConstants.ERR_CODE_400, [CommonErrorCodes.ERROR_INTERNAL_SERVER.code,
                                                CommonErrorCodes.ERROR_INTERNAL_SERVER.messageKey]], False, Exception
            ),
            "OK": ([AppCommonMethod.OK, ], False, [None]),
        },
        "test_get_all_info": {
            "exception": ([CommonConstants.ERR_CODE_400, [CommonErrorCodes.ERROR_INTERNAL_SERVER.code,
                                                          CommonErrorCodes.ERROR_INTERNAL_SERVER.messageKey]],
                          Exception),
            "OK": ([AppCommonMethod.OK, ], [None])
        },
        "test_get_gateway": {
            "exception": (None, [1, ""]),
            "empty": ([], [0, ""]),
            "normal": (["abc"], [0, "abc"]),
        },
        "test_get_eth_ip_map": {
            "exception": (None, [1, ""]),
            "normal": ({"ip": "ip"}, [0, "ip-ip"])
        }
    }

    @staticmethod
    def test_get_eth_ip_map(mocker: MockerFixture, model: GetEthIpMap):
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol", return_value=model.cmd)
        mocker.patch.object(IPV4Checker, "check", return_value=CheckResult(True, ""))
        if model.cmd[0] != 0:
            with pytest.raises(BizException):
                RestoreDefaultsAction.get_eth_ip_map()
        else:
            assert RestoreDefaultsAction.get_eth_ip_map() == model.expect

    @staticmethod
    def test_get_eth_gateway(mocker: MockerFixture, model: GetEthGateway):
        RestoreDefaultsAction.IP_LIST = model.ip
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol", return_value=model.cmd)
        mocker.patch.object(RestoreDefaultsAction, "getgateway", return_value=model.gateway)
        mocker.patch.object(ipaddress, "ip_network").return_value.network_address = model.network
        if model.cmd[0] != 0:
            with pytest.raises(BizException):
                RestoreDefaultsAction.get_eth_gateway()
        else:
            assert RestoreDefaultsAction.get_eth_gateway() == model.expect

    @staticmethod
    def test_get_gateway(mocker: MockerFixture, model: GetGateway):
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol", return_value=model.cmd)
        if model.cmd[0] != 0:
            with pytest.raises(BizException):
                RestoreDefaultsAction.getgateway()
        else:
            assert RestoreDefaultsAction.getgateway() == model.expect

    @staticmethod
    def test_get_all_info(mocker: MockerFixture, model: GetAllInfo):
        mocker.patch.object(RestoreDefaultsAction, "get_eth_ip_map", side_effect=model.map)
        mocker.patch.object(RestoreDefaultsAction, "get_eth_gateway")
        assert RestoreDefaultsAction().get_all_info() == model.expect

    @staticmethod
    def test_post_request(mocker: MockerFixture, model: PostRequest):
        mocker.patch.object(RestoreDefaultsAction, "RESTORE_LOCK").locked.return_value = model.lock
        mocker.patch.object(RestoreDefaultsAction, "check_root_pwd", side_effect=model.check_root)
        mocker.patch.object(RestoreDefaultsAction, "restore_defaults")
        assert RestoreDefaultsAction().post_request({}) == model.expect
