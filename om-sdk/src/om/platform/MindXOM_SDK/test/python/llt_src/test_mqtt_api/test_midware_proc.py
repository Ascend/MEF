from unittest.mock import patch

import mock

from common.file_utils import FileCheck
from fd_msg_process.midware_proc import MidwareProc
from net_manager.manager.fd_cfg_manager import FdCfgManager
from test_mqtt_api.get_log_info import GetLogInfo
from test_mqtt_api.get_log_info import GetOperationLog

getLog = GetLogInfo()
getOplog = GetOperationLog()


class TestMidwareProc:

    @mock.patch.object(FdCfgManager, 'get_cur_fd_ip', mock.Mock(return_value=None))
    @getLog.clear_common_log
    @getOplog.clear_common_log
    def test_handling_msg_from_fd_get_fd_ip_failed(self):
        MidwareProc.handling_msg_from_fd()
        ret = getLog.get_log()
        ret2 = getOplog.get_log()
        assert "get fd ip failed." in ret
        assert "Report event failed." or "No events detected." in ret2

    @mock.patch('time.sleep', mock.Mock(return_value=True))
    @mock.patch.object(MidwareProc, 'report_event', mock.Mock(return_value=True))
    @mock.patch.object(FileCheck, 'check_path_is_exist_and_valid', mock.Mock(return_value=True))
    @mock.patch.object(FdCfgManager, 'get_cur_fd_ip', mock.Mock(return_value=1))
    @getOplog.clear_common_log
    def test_handling_msg_from_fd_succeed(self):
        MidwareProc.handling_msg_from_fd()
        ret = getOplog.get_log()
        assert "Report event successful" in ret

    def test_mock_error(self):
        return [1, "error"]

    @getLog.clear_common_log
    def test_report_event_to_fd(self):
        MidwareProc.report_event_to_fd("a")
        ret = getLog.get_log()
        assert "publish topic:" in ret

    def test_espmanager_event_failed(self):
        return [-1, ""]

    def test_espmanager_event_success(self):
        return [0, ['payload']]

    @getLog.clear_common_log
    def test_report_event_should_failed_when_event_failed(self):
        obj = TestMidwareProc()
        with patch.object(MidwareProc, 'view_functions', wraps={"espmanager/Event": obj.test_espmanager_event_failed}):
            MidwareProc.report_event()
            ret = getLog.get_log()
            assert "report event end" not in ret

    @getLog.clear_common_log
    def test_report_event_should_success_when_event_success(self):
        obj = TestMidwareProc()
        with patch.object(MidwareProc, 'view_functions', wraps={"espmanager/Event": obj.test_espmanager_event_success}):
            MidwareProc.report_event()
            ret = getLog.get_log()
            assert "Report event end" in ret
