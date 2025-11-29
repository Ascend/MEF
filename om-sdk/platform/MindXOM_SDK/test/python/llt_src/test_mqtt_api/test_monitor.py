import mock

from common.file_utils import FileCheck
from common.utils.result_base import Result
from test_mqtt_api.get_log_info import GetLogInfo
from ut_utils.mock_utils import mock_cdll

with mock_cdll():
    from bin.monitor import Monitor

getLog = GetLogInfo()


class TestMonitor:
    @mock.patch.object(FileCheck, 'check_path_is_exist_and_valid', mock.Mock(return_value=Result(False)))
    def test_start_socket_server_should_failed_when_check_failed(self):
        ret = Monitor.start_socket_server()
        assert ret[0] == 1 and ret[1] == "Server start failed."

    def test_start_timer(self):
        ret = Monitor.start_timer()
        assert ret[0] == 0 and ret[1] == "Timer initialize completed."

    def test_get_params(self):
        ret = Monitor.get_params({
            "isLocal": True,
            "action": {
                "McuInfo_all": {
                    "hasList": False
                }
            },
            "description": "Get Mcu Info",
            "enabled": True,
            "intervalTime": 3600,
            "runTimes": 0,
            "minIntervalTime": 1,
            "maxIntervalTime": 86400
        }, None, "description", None)
        assert ret == "Get Mcu Info"


