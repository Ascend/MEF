import json
import os
import shutil
from collections import namedtuple
from typing import NamedTuple

from common.checkers.fd_param_checker import ComputerSystemResetChecker
from common.constants.base_constants import RecoverMiniOSConstants, CommonConstants
from fd_msg_process import fd_common_methods

from fd_msg_process.common_redfish import CommonRedfish
from pytest_mock import MockerFixture

from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from net_manager.manager.fd_cert_manager import FdCertManager
from net_manager.manager.fd_cfg_manager import FdMsgData
from net_manager.manager.net_cfg_manager import NetCfgManager
from net_manager.models import NetManager
from net_manager.schemas import HeaderData, RouteData
from om_fd_msg_process.om_fd_msg_handlers import OMFDMessageHandler

GetDigitalWarranty = namedtuple("GetDigitalWarranty", "expect, data")
DFLCManager = namedtuple("DFLCManager", "expect, data, lib_interface")
HandleMsg = namedtuple("HandleMsg", "config")
Reset = namedtuple("Reset", "expect, data, check_param, locked, lib_interface")


class RecoverMiniOS(NamedTuple):
    expect: list[int, dict] = [
        -1,
        {
            'percentage': '0%',
            'reason': 'Recover mini os failed',
            'result': 'failed',
            'topic': 'min_recovery'
        }
    ]
    locked: bool = False
    ret_dict: dict = {"status": 200, "message": {"system_busy": False}}
    ip: str = ""
    is_exception: bool = False


class TestOMFDMessageHandler:
    use_cases = {
        "test_check_external_param": {
            "failed": (-1, {"restart_method": "abc"}),
            "normal": (0, {"restart_method": "Graceful"}),
        },
        "test_config_netmanager_dflc": {
            "data not dict": ([-1, "FAILED: wrong param type"], [], None),
            "life_span not in data": ([-1, "FAILED: need config start_point and life_span at same time"],
                                      {"start_point": "1"}, None),
            "data wrong": ([-1, "FAILED: input param check failed"], {"start_point": "1", "life_span": 1}, None),
            "lib failed": ([-1, "FAILED: set dflc info failed"], {"start_point": "2022-09-09", "life_span": 1},
                           {"status": 400, "message": {}}),
            "lib exception": ([-1, "FAILED: set dflc info failed"], {"start_point": "2022-09-09", "life_span": 1},
                              Exception()),
            "normal": ([0, "SUCCESS"], {"start_point": "2022-09-09", "life_span": 1},
                       [{"status": 200, "message": {}}]),
        },
        "test_handle_msg_config_dflc": {
            "config failed": ([],),
            "normal": ([0, "success"],),
        },
        "test_computer_system_reset": {
            "param invalid": ([-1, "err"], {"restart_method": "abc"}, [-1, "err"], None, None),
            "restart method null": ([-1, {"restartable": "false", "reason": "ERR.603, Input parameter invalid!"}],
                                    {}, [0, ""], None, None),
            "restart method wrong": ([-1, {"restartable": "false", "reason": "ERR.603, Parameter reset_type error."}],
                                     {"restart_method": "Reset"}, [0, ""], None, None),
            "lock and graceful": ([-1, {"restartable": "false", "reason": "ERR.604, Resource is busy"}],
                                  {"restart_method": "Graceful"}, [0, ""], True, None),
            "lock and force failed": ([-1, {"restartable": "true", "reason": "ERR.600, err"}],
                                      {"restart_method": "Force"}, [0, ""], True,
                                      [{"status": 400, "message": "err"}]),
            "lock and force success": ([0, {"restartable": "true", "reason": "resource is busy"}],
                                       {"restart_method": "Force"}, [0, ""], True,
                                       [{"status": 200, "message": {}}]),
            "normal failed": ([-1, {"restartable": "false", "reason": "ERR.600, err"}],
                              {"restart_method": "Force"}, [0, ""], False,
                              [{"status": 400, "message": "err"}]),
            "normal success": ([0, {"restartable": "true", "reason": "System is restartable"}],
                               {"restart_method": "Graceful"}, [0, ""], False,
                               [{"status": 200, "message": ""}]),
            "exception": ([-1, {"restartable": "true", "reason": "ERR.600, System restart error."}],
                          {"restart_method": "ColdReset"}, [0, ""], False,
                          Exception()),
        },
        "test_recover_mini_os": {
            "locked": RecoverMiniOS(locked=True),
            "exclusive": RecoverMiniOS(ret_dict={"status": 400, "message": "fail"}),
            "ret dict invalid": RecoverMiniOS(ret_dict={"status": 200, "message": "success"}),
            "success": RecoverMiniOS(
                expect=[
                    0,
                    {
                        'topic': 'min_recovery',
                        'percentage': '100%',
                        'result': 'success',
                        'reason': 'Prepare for recover mini os successfully.'
                    }
                ],
                ip="127.0.0.1"
            ),
            "recover error": RecoverMiniOS(
                expect=[
                    -1,
                    {
                        'topic': 'min_recovery',
                        'percentage': '0%',
                        'result': 'failed',
                        'reason': 'Prepare for recover mini os has exception'
                    }
                ]
            ),
            "exception error": RecoverMiniOS(
                expect=[
                    -1,
                    {
                        'topic': 'min_recovery',
                        'percentage': '0%',
                        'result': 'failed',
                        'reason': 'Prepare for recover mini os has internal exception'
                    }
                ],
                ip="127.0.0.1",
                is_exception=True
            )
        }
    }

    @staticmethod
    def test_check_external_param(model: GetDigitalWarranty):
        payload_publish = {
            "restartable": "",
            "result": "",
            "reason": ""
        }
        ret = OMFDMessageHandler.check_external_param(ComputerSystemResetChecker, model.data, payload_publish, "err %s")
        assert ret[0] == model.expect

    @staticmethod
    def test_config_netmanager_dflc(mocker: MockerFixture, model: DFLCManager):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        ret = OMFDMessageHandler.config_netmanager_dflc(model.data)
        assert ret == model.expect

    @staticmethod
    def test_handle_msg_config_dflc(mocker: MockerFixture, model: HandleMsg):
        mocker.patch.object(OMFDMessageHandler, "config_netmanager_dflc", return_value=model.config)
        mocker.patch.object(fd_common_methods, "publish_ws_msg")
        msg = FdMsgData(HeaderData("1", "1", False, 1), RouteData("1", "1", "1", "1"), {})
        ret = OMFDMessageHandler.handle_msg_config_dflc(msg)
        assert ret is None

    @staticmethod
    def test_computer_system_reset(mocker: MockerFixture, model: Reset):
        mocker.patch.object(OMFDMessageHandler, "check_external_param", return_value=model.check_param)
        mocker.patch.object(CommonRedfish, "SYS_CRITIC_LOCK").locked.return_value = model.locked
        mocker.patch.object(fd_common_methods, "publish_ws_msg")
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        ret = OMFDMessageHandler.computer_system_reset(model.data)
        assert ret == model.expect

    @staticmethod
    def test_handle_computer_system_reset_msg_from_fd_by_mqtt(mocker: MockerFixture):
        mocker.patch.object(OMFDMessageHandler, "computer_system_reset")
        mocker.patch.object(FdMsgData, "gen_ws_msg_obj")
        mocker.patch.object(fd_common_methods, "publish_ws_msg")
        msg = FdMsgData(HeaderData("1", "1", False, 1), RouteData("1", "1", "1", "1"), {})
        ret = OMFDMessageHandler.handle_computer_system_reset_msg_from_fd_by_mqtt(msg)
        assert ret is None

    @staticmethod
    def test_recover_mini_os(mocker: MockerFixture, model: RecoverMiniOS, monkeypatch):
        if model.locked:
            mocker.patch.object(RedfishGlobals, "high_risk_exclusive_lock").locked().return_value = model.locked

        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface").return_value = model.ret_dict
        mini_os_flag = "/tmp/recover_mini_os_flag"
        mocker.patch.object(RecoverMiniOSConstants, "RECOVER_FLAG", mini_os_flag)
        net_manager = NetManager()
        net_manager.net_mgmt_type = "FusionDirector"
        net_manager.ip = model.ip
        net_manager.port = "443"
        mocker.patch.object(NetCfgManager, "get_net_cfg_info").return_value = net_manager
        mocker.patch.object(FdCertManager, "cert_for_restore_mini_os").return_value = {
            "name": "name",
            "source": "source",
            "update_time": "2024-03-05",
            "cert_contents": "xxx",
            "crl_contents": "xxx",
        }
        tmp_redfish = "/tmp/redfish"
        tmp_config = "/tmp/config"
        for dir_path in tmp_redfish, tmp_config:
            if not os.path.exists(dir_path):
                os.makedirs(dir_path)

        certs = ("redfish_encrypt.keystore", "redfish_encrypt_backup.keystore", "om_alg.json")
        mode = os.O_WRONLY | os.O_CREAT | os.O_TRUNC
        for file in certs:
            with os.fdopen(os.open(os.path.join(tmp_redfish, file), mode, 0o600), "w") as fd:
                fd.write("test")

        mocker.patch.object(CommonConstants, "CONFIG_HOME_PATH", "/tmp")
        mocker.patch.object(RecoverMiniOSConstants, "CONFIG_PATH", tmp_config)
        if model.is_exception:
            monkeypatch.setattr(json, "dumps", 1111)

        ret = OMFDMessageHandler().recover_mini_os()
        assert ret == model.expect
        if os.path.exists(mini_os_flag):
            os.unlink(mini_os_flag)
            
        for dir_path in tmp_redfish, tmp_config:
            if os.path.exists(dir_path):
                shutil.rmtree(dir_path)

    @staticmethod
    def test_handle_recover_mini_os(mocker):
        mocker.patch.object(OMFDMessageHandler, "recover_mini_os")
        mocker.patch.object(FdMsgData, "gen_ws_msg_obj")
        msg = FdMsgData(HeaderData("1", "1", False, 1), RouteData("1", "1", "1", "1"), {})
        ret = OMFDMessageHandler.handle_recover_mini_os(msg)
        assert ret is None
