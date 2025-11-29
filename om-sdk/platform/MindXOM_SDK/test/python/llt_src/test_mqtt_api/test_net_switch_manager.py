# Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import copy
import mock
import pytest

from common.constants.base_constants import CommonConstants
from common.utils.result_base import Result
from common.schema import AdapterResult
from common.utils.app_common_method import AppCommonMethod
from lib_restful_adapter import LibRESTfulAdapter
from mef_msg_process.mef_proc import MefProc
from net_manager.manager.fd_cfg_manager import FdConfigData, FdCfgManager
from net_manager.manager.net_switch_manager import WebNetSwitchManager, FdNetSwitchManager
from net_manager.manager.net_cfg_manager import NetCfgManager
from net_manager.models import NetManager
from net_manager.schemas import SystemInfo
from test_mqtt_api.get_log_info import GetLogInfo
from wsclient.connect_status import FdConnectStatus
from wsclient.fd_connect_check import FdConnectCheck
from wsclient.ws_client_mef import WsClientMef
from wsclient.ws_client_mgr import WsClientMgr
from wsclient.ws_monitor import WsMonitor


getLog = GetLogInfo()


class TestWebNetSwitchManager:
    @staticmethod
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager()))
    def test_init():
        web_mode = WebNetSwitchManager(dict())
        assert isinstance(web_mode.net_cfg_manager, NetCfgManager)

    @staticmethod
    def test_update_host_info():
        with pytest.raises(Exception) as exception_info:
            WebNetSwitchManager.update_host_info({"operate_type": "clear", "fd_server_name": "server_name"})
            assert "Update host info failed" in str(exception_info.value)

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager(server_name="server name")))
    @mock.patch.object(WebNetSwitchManager, "update_host_info", mock.Mock(return_value=True))
    @mock.patch.object(WebNetSwitchManager, "update_net_info", mock.Mock(return_value=True))
    @mock.patch.object(WsMonitor, "stop_all_connect_threads", mock.Mock(return_value=True))
    @mock.patch.object(WsClientMef, "stop_mef_process", mock.Mock(return_value=True))
    @mock.patch.object(FdConnectStatus, "trans_to_not_configured", mock.Mock(return_value=True))
    @mock.patch.object(FdCfgManager, "modify_alarm", mock.Mock(return_value=True))
    def test_switch_deal():
        ret = WebNetSwitchManager(dict()).switch_deal()
        assert getLog.get_log() is not None and ret[0] == 0


class TestFdNetSwitchManager:
    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(copy, "deepcopy", mock.Mock(return_value=dict()))
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager()))
    @mock.patch.object(NetManager, "from_dict", mock.Mock(return_value=NetManager()))
    @mock.patch.object(FdNetSwitchManager, "get_sys_info", mock.Mock(return_value=SystemInfo(asset_tag="asset tag")))
    @mock.patch.object(FdConfigData, "from_dict", mock.Mock(return_value=FdConfigData()))
    @mock.patch.object(FdConnectCheck, "connect_test", mock.Mock(return_value=Result(False)))
    def test_connect_test_failed():
        assert getLog.get_log() is not None and not FdNetSwitchManager(dict()).connect_test()

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager()))
    @mock.patch.object(copy, "deepcopy", mock.Mock(return_value=dict()))
    @mock.patch.object(NetManager, "from_dict", mock.Mock(return_value=NetManager()))
    @mock.patch.object(FdNetSwitchManager, "get_sys_info", mock.Mock(return_value=SystemInfo(asset_tag="asset tag")))
    @mock.patch.object(FdConfigData, "from_dict", mock.Mock(return_value=FdConfigData()))
    @mock.patch.object(FdConnectCheck, "connect_test", mock.Mock(return_value=Result(True)))
    def test_connect_test_success():
        assert getLog.get_log() is not None and FdNetSwitchManager(dict()).connect_test()

    @staticmethod
    def test_get_sys_info_exception():
        with pytest.raises(Exception) as exception_info:
            FdNetSwitchManager(dict()).get_sys_info()
            assert "Get system info failed" in str(exception_info.value)

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager()))
    @mock.patch.object(AdapterResult, "from_dict", mock.Mock(
        return_value=AdapterResult(status=AppCommonMethod.OK,
                                   message={"AssetTag": "asset tag", "Model": "product_name"})))
    def test_get_sys_info_without_serial_number():
        ret = FdNetSwitchManager({"NodeId": 1}).get_sys_info()
        assert getLog.get_log() is not None and ret.asset_tag == "asset tag" and ret.serial_number == 1

    @staticmethod
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager()))
    @mock.patch.object(AdapterResult, "from_dict", mock.Mock(
        return_value=AdapterResult(status=AppCommonMethod.OK,
                                   message={"AssetTag": "asset tag", "Model": "product_name", "SerialNumber": 1})))
    def test_get_sys_info_without_serial_number():
        ret = FdNetSwitchManager(dict()).get_sys_info()
        assert ret.asset_tag == "asset tag" and ret.serial_number == 1

    @staticmethod
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager(server_name="server name")))
    def test_get_net_info_with_rollback():
        ret = FdNetSwitchManager(dict()).get_net_info()
        assert ret.server_name == "server name"

    @staticmethod
    @mock.patch.object(NetCfgManager, "get_net_cfg_info",
                       mock.Mock(return_value=NetManager(ip="1.1.1.1", cloud_user="admin",
                                                         cloud_pwd="123", status="connecting")))
    @mock.patch.object(NetManager, "encrypt_cloud_pwd", mock.Mock(return_value=""))
    def test_get_net_info_with_manual_update_pwd():
        ret = FdNetSwitchManager({"NetIP": "1.1.1.1", "NetAccount": "admin", "NetPassword": ""}).get_net_info(False)
        assert ret.status == "connecting"

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager()))
    @mock.patch.object(WsClientMgr, "clear_connect_result", mock.Mock(return_value=True))
    def test_switch_deal_with_docker_root_not_mounted():
        ret = FdNetSwitchManager({"NetIP": "1.1.1.1", "NetAccount": "admin", "NetPassword": ""}).switch_deal()
        assert getLog.get_log() is not None and ret[0] == 110226

    @staticmethod
    @mock.patch.object(WsClientMgr, "clear_connect_result", mock.Mock(return_value=True))
    @mock.patch.object(LibRESTfulAdapter, "lib_restful_interface",
                       mock.Mock(return_value={"message": {"SupportModel": CommonConstants.ATLAS_A500_A2}}))
    def test_switch_deal_with_connect_test_exception():
        with pytest.raises(Exception) as exception_info:
            ret = FdNetSwitchManager({"test": True}).switch_deal()
            assert "Switch FusionDirector manage failed" in str(exception_info.value)

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager()))
    @mock.patch.object(WsClientMgr, "clear_connect_result", mock.Mock(return_value=True))
    @mock.patch.object(LibRESTfulAdapter, "lib_restful_interface",
                       mock.Mock(return_value={"message": {"SupportModel": CommonConstants.ATLAS_A500_A2}}))
    @mock.patch.object(FdNetSwitchManager, "connect_test", mock.Mock(return_value=False))
    @mock.patch.object(WsClientMgr, "check_account_info_invalid", mock.Mock(return_value=True))
    def test_switch_deal_with_check_account_info_invalid_true():
        ret = FdNetSwitchManager({"test": True}).switch_deal()
        assert getLog.get_log() is not None and ret[0] == 110207

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(WsClientMgr, "clear_connect_result", mock.Mock(return_value=True))
    @mock.patch.object(LibRESTfulAdapter, "lib_restful_interface",
                       mock.Mock(return_value={"message": {"SupportModel": CommonConstants.ATLAS_A500_A2}}))
    @mock.patch.object(FdNetSwitchManager, "connect_test", mock.Mock(return_value=False))
    @mock.patch.object(WsClientMgr, "check_account_info_invalid", mock.Mock(return_value=False))
    @mock.patch.object(WsClientMgr, "check_ip_locked", mock.Mock(return_value=True))
    def test_switch_deal_with_check_ip_locked_true():
        ret = FdNetSwitchManager({"test": True}).switch_deal()
        assert getLog.get_log() is not None and ret[0] == 110225

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(WsClientMgr, "clear_connect_result", mock.Mock(return_value=True))
    @mock.patch.object(LibRESTfulAdapter, "lib_restful_interface",
                       mock.Mock(return_value={"message": {"SupportModel": CommonConstants.ATLAS_A500_A2}}))
    @mock.patch.object(FdNetSwitchManager, "connect_test", mock.Mock(return_value=False))
    @mock.patch.object(WsClientMgr, "check_account_info_invalid", mock.Mock(return_value=False))
    @mock.patch.object(WsClientMgr, "check_ip_locked", mock.Mock(return_value=False))
    def test_switch_deal_with_check_ip_locked_true():
        with pytest.raises(Exception) as exception_info:
            ret = FdNetSwitchManager({"test": True}).switch_deal()
            assert "Switch FusionDirector manage failed" in str(exception_info.value)

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(WsClientMgr, "clear_connect_result", mock.Mock(return_value=True))
    @mock.patch.object(LibRESTfulAdapter, "lib_restful_interface",
                       mock.Mock(return_value={"message": {"SupportModel": CommonConstants.ATLAS_A500_A2}}))
    @mock.patch.object(FdNetSwitchManager, "update_net_info", mock.Mock(return_value=True))
    @mock.patch.object(WsMonitor, "stop_all_connect_threads", mock.Mock(return_value=True))
    @mock.patch.object(FdConnectStatus, "trans_to_connecting", mock.Mock(return_value=True))
    @mock.patch.object(WsMonitor, "start_fd_connect_monitor", mock.Mock(return_value=True))
    @mock.patch.object(MefProc, "start_mef_connect_timer", mock.Mock(return_value=True))
    @mock.patch.object(NetManager, "encrypt_cloud_pwd", mock.Mock(return_value=""))
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager(status="ready")))
    @mock.patch.object(WsClientMgr, "ready_for_send_msg", mock.Mock(return_value=True))
    def test_switch_deal_with_not_test_and_switch_success():
        ret = FdNetSwitchManager({"test": False}).switch_deal()
        assert getLog.get_log() is not None and ret[0] == 0

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(WsClientMgr, "clear_connect_result", mock.Mock(return_value=True))
    @mock.patch.object(LibRESTfulAdapter, "lib_restful_interface",
                       mock.Mock(return_value={"message": {"SupportModel": CommonConstants.ATLAS_A500_A2}}))
    @mock.patch.object(FdNetSwitchManager, "update_net_info", mock.Mock(return_value=True))
    @mock.patch.object(WsMonitor, "stop_all_connect_threads", mock.Mock(return_value=True))
    @mock.patch.object(FdConnectStatus, "trans_to_connecting", mock.Mock(return_value=True))
    @mock.patch.object(WsMonitor, "start_fd_connect_monitor", mock.Mock(return_value=True))
    @mock.patch.object(MefProc, "start_mef_connect_timer", mock.Mock(return_value=True))
    @mock.patch.object(NetManager, "encrypt_cloud_pwd", mock.Mock(return_value=""))
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager(status="connecting")))
    @mock.patch.object(WsClientMgr, "check_account_info_invalid", mock.Mock(return_value=True))
    def test_switch_deal_with_not_test_and_check_account_failed():
        ret = FdNetSwitchManager({"test": False}).switch_deal()
        assert getLog.get_log() is not None and ret[0] == 110207

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(WsClientMgr, "clear_connect_result", mock.Mock(return_value=True))
    @mock.patch.object(LibRESTfulAdapter, "lib_restful_interface",
                       mock.Mock(return_value={"message": {"SupportModel": CommonConstants.ATLAS_A500_A2}}))
    @mock.patch.object(FdNetSwitchManager, "update_net_info", mock.Mock(return_value=True))
    @mock.patch.object(WsMonitor, "stop_all_connect_threads", mock.Mock(return_value=True))
    @mock.patch.object(FdConnectStatus, "trans_to_connecting", mock.Mock(return_value=True))
    @mock.patch.object(WsMonitor, "start_fd_connect_monitor", mock.Mock(return_value=True))
    @mock.patch.object(MefProc, "start_mef_connect_timer", mock.Mock(return_value=True))
    @mock.patch.object(NetManager, "encrypt_cloud_pwd", mock.Mock(return_value=""))
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager(status="connecting")))
    @mock.patch.object(WsClientMgr, "check_account_info_invalid", mock.Mock(return_value=False))
    @mock.patch.object(WsClientMgr, "check_ip_locked", mock.Mock(return_value=True))
    def test_switch_deal_with_not_test_and_ip_locked():
        ret = FdNetSwitchManager({"test": False}).switch_deal()
        assert getLog.get_log() is not None and ret[0] == 110225

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(WsClientMgr, "clear_connect_result", mock.Mock(return_value=True))
    @mock.patch.object(LibRESTfulAdapter, "lib_restful_interface",
                       mock.Mock(return_value={"message": {"SupportModel": CommonConstants.ATLAS_A500_A2}}))
    @mock.patch.object(FdNetSwitchManager, "update_net_info", mock.Mock(return_value=True))
    @mock.patch.object(WsMonitor, "stop_all_connect_threads", mock.Mock(return_value=True))
    @mock.patch.object(FdConnectStatus, "trans_to_connecting", mock.Mock(return_value=True))
    @mock.patch.object(WsMonitor, "start_fd_connect_monitor", mock.Mock(return_value=True))
    @mock.patch.object(MefProc, "start_mef_connect_timer", mock.Mock(return_value=True))
    @mock.patch.object(NetManager, "encrypt_cloud_pwd", mock.Mock(return_value=""))
    @mock.patch.object(NetCfgManager, "get_net_cfg_info", mock.Mock(return_value=NetManager(status="connecting")))
    @mock.patch.object(WsClientMgr, "check_account_info_invalid", mock.Mock(return_value=False))
    @mock.patch.object(WsClientMgr, "check_ip_locked", mock.Mock(return_value=False))
    @mock.patch.object(WsClientMgr, "check_cert_is_invalid", mock.Mock(return_value=True))
    def test_switch_deal_with_not_test_and_success():
        ret = FdNetSwitchManager({"test": False}).switch_deal()
        assert getLog.get_log() is not None and ret[0] == 206
