from unittest import mock

from fd_msg_process.fd_msg_handlers import FDMessageHandler
from fd_msg_process.midware_route import MidwareRoute
from common.utils.result_base import Result
from net_manager.manager.fd_cfg_manager import FdMsgData
from test_mqtt_api.get_log_info import GetLogInfo

getLog = GetLogInfo()


class TestFDHandlers:

    def test_mock_run(self):
        return [0, ""]

    @mock.patch.object(FDMessageHandler, 'call_view_func', mock.Mock(return_value=True))
    @getLog.clear_common_log
    def test_handle_msg_from_fd_by_mqtt_topic_three_should_ok(self):
        msg = FdMsgData.gen_ws_msg_obj({"restart_method": "Graceful"}, "websocket/restart")
        FDMessageHandler.handle_msg_from_fd_by_mqtt(msg)
        ret = getLog.get_log()
        assert "Receive topic" in ret
        assert "Done topic" in ret

    @mock.patch.object(FDMessageHandler, 'call_view_func', mock.Mock(return_value=True))
    @getLog.clear_common_log
    def test_handle_reset_alarm_from_fd(self):
        msg = FdMsgData.gen_ws_msg_obj({"resource": "alarm"}, "websocket/rearm")
        FDMessageHandler.handle_reset_alarm_from_fd(msg)
        ret = getLog.get_log()
        assert "Receive topic" in ret
        assert "Done topic" in ret

    @mock.patch.object(FDMessageHandler, 'call_view_func', mock.Mock(return_value=Result(result=True, data=[0, ""])))
    @getLog.clear_common_log
    def test_handle_msg_from_fd_by_mqtt_topic_one_should_ok(self):
        msg = FdMsgData.gen_ws_msg_obj({"hostname": "g000121002000220010-ipc01"}, "websocket/config_hostname")
        msg.deal_func_name = "espmanager/Hostname"
        msg.up_resource = "websocket/config_result"
        FDMessageHandler.handle_msg_from_fd_by_mqtt(msg)
        ret = getLog.get_log()
        assert "Receive topic" in ret
        assert "Done topic" in ret

    @mock.patch.object(FDMessageHandler, 'call_view_func', mock.Mock(return_value=False))
    @getLog.clear_common_log
    def test_handle_msg_from_fd_by_mqtt_topic_two_should_failed(self):
        msg = FdMsgData.gen_ws_msg_obj({"hostname": "g000121002000220010-ipc01"}, "websocket/config_hostname")
        FDMessageHandler.handle_msg_from_fd_by_mqtt(msg)
        ret = getLog.get_log()
        assert "Receive topic" in ret
        assert "result of deal func is invalid" in ret

    @mock.patch.object(MidwareRoute, 'view_functions', {'a': test_mock_run})
    def test_call_view_func_should_return_true(self):
        ret = FDMessageHandler.call_view_func("a", "")
        assert ret

    def test_call_view_func_should_return_false(self):
        ret = FDMessageHandler.call_view_func("a", "")
        assert not ret
