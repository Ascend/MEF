import threading

import pytest
from pytest_mock import MockerFixture

from common.utils.timer import RepeatingTimer
from fd_msg_process.midware_proc import MidwareProc
from fd_msg_process.midware_urls import MidwareUris
from mef_msg_process.mef_proc import MefProc
from net_manager.manager.fd_cfg_manager import FdCfgManager
from net_manager.manager.net_cfg_manager import NetCfgManager
from test_mqtt_api.get_log_info import GetLogInfo

from ibma_redfish_urls import RedfishURIs
from ibma_redfish_main import RedfishMain, LibRESTfulAdapter, RedfishGlobals, UploadMarkFile
from wsclient.ws_monitor import WsMonitor

getLog = GetLogInfo()


class TestRedfishMain:
    use_cases = {

    }

    def test_rf_start_server(self, mocker: MockerFixture):
        mocker.patch.object(RedfishURIs, "rf_api_ibma_server")
        ret = RedfishMain.rf_start_server("50.50.50.50", 80, "ResourceDefV1")
        assert ret is None

    def test_rf_start_server_exception(self):
        with pytest.raises(ValueError):
            RedfishMain.rf_start_server("50.50.50.50", 80, "ResourceDefV2")

    def test_stop_handler(self, mocker: MockerFixture):
        mocker.patch("os.kill")
        mocker.patch("os.getpid")
        ret = RedfishMain.stop_handler()
        assert ret is None

    def test_start_monitor_timer(self, mocker: MockerFixture):
        mocker.patch.object(LibRESTfulAdapter, "start_timer", return_value=[0, "success"])
        ret = RedfishMain.start_monitor_timer()
        assert ret is None

    def test_main(self, mocker: MockerFixture):
        mocker.patch("signal.signal")
        mocker.patch.object(MidwareUris, "mid_ware_add_route")
        mocker.patch.object(threading, "Thread")
        mocker.patch.object(threading, "Timer")
        mocker.patch("ibma_redfish_main.start_keystore_update_tasks")
        mocker.patch.object(RedfishMain, "mid_ware_task")
        mocker.patch.object(RedfishMain, "start_database_monitor_task")
        mocker.patch.object(RedfishMain, "start_monitor_timer")
        mocker.patch.object(MidwareProc, "dispatch_fd_messages")
        mocker.patch.object(RedfishGlobals, "init_http_server_param")
        mocker.patch.object(RedfishGlobals, "get_http_port")
        mocker.patch.object(RedfishMain, "config_db")
        mocker.patch("ibma_redfish_main.RepeatingTimer")
        mocker.patch.object(UploadMarkFile, "clear_upload_mark_file_all")
        mocker.patch.object(RedfishMain, "start_check_user_lock")
        mocker.patch.object(NetCfgManager, "get_net_cfg_info")
        mocker.patch.object(FdCfgManager, "check_and_reset_status")
        mocker.patch.object(FdCfgManager, "start_cert_status_monitor")
        mocker.patch.object(WsMonitor, "start_fd_connect_monitor")
        mocker.patch.object(RedfishMain, "init_ability")
        mocker.patch.object(MefProc, "start_mef_connect_timer")
        mocker.patch.object(RedfishMain, "rf_start_server")
        mocker.patch("sys.stdout.flush")
        ret = RedfishMain.main()
        assert ret is None

    @getLog.clear_common_log
    def test_do_work(self):
        RedfishMain.do_work(["A"])
        ret = getLog.get_log()
        assert "A not found in view_functions" in ret
