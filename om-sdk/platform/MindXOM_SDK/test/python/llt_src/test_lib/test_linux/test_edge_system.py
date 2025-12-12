import socket
from collections import namedtuple

from pytest_mock import MockerFixture

from bin.monitor_config import SystemSetting
from common.file_utils import FileCheck
from common.utils.exec_cmd import ExecCmd
from common.utils.system_utils import SystemUtils
from common.utils.result_base import Result
from common.common_methods import CommonMethods
from ut_utils.mock_utils import mock_cdll, mock_write_file_with_os_open, mock_read_data, mock_npu_smi_board_type, \
    mock_npu_smi_npu_version

with mock_cdll():
    with mock_npu_smi_board_type():
        with mock_npu_smi_npu_version():
            from lib.Linux.systems.systems import SystemInfo
            from devm.device_mgr import DEVM, Device
            from devm.exception import DeviceManagerError

CheckRequestData = namedtuple("CheckRequestData", "expect, request")
ELabelCheck = namedtuple("ELabelCheck", "expect, info")
PatchRequest = namedtuple("PatchRequest", "expect, request, lock, tags, mcu, offset")
SetDatetimeLocalOffset = namedtuple("SetDatetimeLocalOffset", "expect, cmd")
GetTimeZone = namedtuple("GetTimeZone", "expect, cmd")
GetPcbID = namedtuple("GetPcbID", "expect, pcb")
GetSysUptime = namedtuple("GetSysUptime", "expect, check, time")
GetSysTime = namedtuple("GetSysTime", "expect, os_time")
GetCpuUsage = namedtuple("GetCpuUsage", "expect, content, stat")
GetCpuStatInfo = namedtuple("GetCpuStatInfo", "expect, info_list, logical")
GetOsVersionInfo = namedtuple("GetOsVersionInfo", "expect, read")
GetKernelNumber = namedtuple("GetKernelNumber", "expect, version")
GetSysElectronicTags = namedtuple("GetSysElectronicTags", "expect, sys")
GetHostName = namedtuple("GetHostName", "expect, name")
GetSysUuid = namedtuple("GetSysUuid", "expect, uuid")
GetModel = namedtuple("GetModel", "expect, path_valid, read, get_prd_name")
GetMcuTemperature = namedtuple("GetMcuTemperature", "expect, board")
GetMcuVoltage = namedtuple("GetMcuVoltage", "expect, voltage")
GetPower = namedtuple("GetPower", "expect, power")
GetMinidTemperature = namedtuple("GetMinidTemperature", "expect, npu, tmp")
GetKernelVersion = namedtuple("GetKernelVersion", "expect, out")
SetTime = namedtuple("SetTime", "expect, time, set_os_time")
FormatTime = namedtuple("SetTime", "expect, time")
SetOSTime = namedtuple("SetTime", "expect, cmd")


class TestSystem:
    use_cases = {
        "test_check_request_data": {
            "null request": (False, None),
            "empty request": (False, ""),
            "valid request": (True, {})
        },
        "test_patch_request": {
            "busy": ([CommonMethods.ERROR, "System modify is busy"], None, True, None, None, None),
            "wrong param": ([CommonMethods.ERROR, "Parameter is wrong"], {"any": ""}, False, None, None, None),
            "empty asset_tag": (
                [CommonMethods.ERROR, "ERR.0155,Failed to set electronic tags."], {"AssetTag": ""}, False, None, None,
                None
            ),
            "Set electronic failed": (
                [CommonMethods.ERROR, "Set electronic failed"], {"AssetTag": "a"}, False,
                [CommonMethods.ERROR, "Set electronic failed"], None, None
            ),
            "empty date_time_local_offset": (
                [CommonMethods.ERROR, "ERR.01004,Set timezones failed"], {"DateTimeLocalOffset": ""}, False, None, None,
                None
            ),
            "Set time zone failed.": (
                [CommonMethods.ERROR, "Set time zone failed."], {"DateTimeLocalOffset": "abc"}, False, None, None,
                [CommonMethods.ERROR, "Set time zone failed."]
            ),
            "OK": (
                [CommonMethods.OK, ], {"HostName": "hostname"}, False, None, None, None
            )
        },
        "test_set_date_time_local_offset": {
            "error": ([400, "ERR.01004,Set timezones failed"], 1),
            "ok": ([CommonMethods.OK, "ok"], 0),
        },
        "test_get_pcd_id": {
            "normal": (1, 1),
        },
        "test_get_sys_uptime": {
            "check failed": (None, Result(False, "failed"), None),
            "success": ("00:00:33 10 days", Result(True), "864033.30"),
        },
        "test_get_sys_time": {
            "normal": ("a b c", "a b c"),
        },
        "test_get_cpu_usage": {
            "read content failed": (-1, [None, None], [None, None]),
            "exception": (-1, ["ret", "ret"], [None, None]),
            "normal": (100, ["ret", "ret"], [[0, 0], [0, 100]]),
            "zero total": (-1, ["ret", "ret"], [[0, 0], [0, 0]]),
        },
        "test_get_cpu_stat_info": {
            "not startswith cpu": ([], [""], []),
            "length ne 10": ([], ["cpu  "], []),
            "success": ([1, 8], ["cpu " + " 1" * 10], []),
            "invalid": ([], ["cpu " + " a" * 10], []),
            "invalid logical": ([], [], ["a"]),
        },
        "test_get_os_version_info": {
            "read version failed": ("", Exception),
            "normal": ("", ['PRETTY_NAME="Euler"\nVERSION_ID="22.04"']),
        },
        "test_get_kernel_number": {
            "null": (None, None),
            "normal": ("version", "version"),
        },
        "test_get_sys_electronic_tags": {
            "normal": (1, 1),
        },
        "test_get_host_name": {
            "normal": ("name", "name")
        },
        "test_get_sys_uuid": {
            "normal": ("tag", "tag")
        },
        "test_get_mcu_temperature": {
            "normal": ("tmp", "tmp")
        },
        "test_get_mcu_voltage": {
            "normal": (1.00, 100)
        },
        "test_get_power": {
            "normal": (2.00, 2)
        },
        "test_get_minid_temperature": {
            "normal": (1, ["npu0"], 1),
        },
        "test_get_kernel_version": {
            "normal": ("4.19.90", "4.19.90")
        },
        "test_set_time": {
            "param_is_wrong": ([400, "ERR.0141,Failed to Setting mcu time."], "", None),
            "failed": ([500, "Set OS time failed."], "2023-01-12",
                       [400, "ERR.0141,Failed to Setting os time."]),
            "success": ([CommonMethods.OK, ""], "2023-01-12", [CommonMethods.OK, "Set MCU time successfully."])
        },
        "test_format_time": {
            "param_is_wrong": ([400, "format time failed"], ""),
            "success": ([CommonMethods.OK, "2023-01-12"], "2023-01-12")
        },
        "test_set_os_time": {
            "failed": ([CommonMethods.INTERNAL_ERROR, "Set system time failed."], [-1000, 'call linux command error']),
            "success": ([CommonMethods.OK, "set system time success."], [0, ""])
        },
    }

    @staticmethod
    def test_check_request_data(model: CheckRequestData):
        assert SystemInfo.check_request_data(model.request) == model.expect

    @staticmethod
    def test_patch_request(mocker: MockerFixture, model: PatchRequest):
        mocker.patch.object(SystemInfo, "SYSTEM_LOCK").locked.return_value = model.lock
        mocker.patch.object(SystemInfo, "set_sys_electronic_tags", return_value=model.tags)
        mocker.patch.object(SystemInfo, "set_date_time_local_offset", return_value=model.offset)
        mocker.patch.object(SystemInfo, "save_time_zone_config")
        mocker.patch.object(SystemInfo, "get_all_system_info")
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=0)
        mock_write_file_with_os_open(mocker)
        assert SystemInfo().patch_request(model.request) == model.expect

    @staticmethod
    def test_set_date_time_local_offset(mocker: MockerFixture, model: SetDatetimeLocalOffset):
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=model.cmd)
        mocker.patch("time.tzset")
        mocker.patch("time.sleep")
        assert SystemInfo().set_date_time_local_offset("") == model.expect

    @staticmethod
    def test_get_pcd_id(mocker: MockerFixture, model: GetPcbID):
        mocker.patch.object(Device, "get_attribute", return_value=model.pcb)
        info = SystemInfo()
        info.main_board = Device
        info.get_pcb_id()
        assert info.PCBVersion == model.expect

    @staticmethod
    def test_get_sys_uptime(mocker: MockerFixture, model: GetSysUptime):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check)
        mock_read_data(mocker, model.time)
        info = SystemInfo()
        info.get_sys_uptime()
        assert info.Uptime == model.expect

    @staticmethod
    def test_get_sys_time(mocker: MockerFixture, model: GetSysTime):
        mocker.patch("time.strftime", return_value=model.os_time)
        system = SystemInfo()
        system.get_sys_time()
        assert system.Datetime == model.expect

    @staticmethod
    def test_get_cpu_usage(mocker: MockerFixture, model: GetCpuUsage):
        mocker.patch.object(SystemInfo, "get_file_content", side_effect=model.content)
        mocker.patch("time.sleep")
        mocker.patch.object(SystemInfo, "get_cpu_stat_info", side_effect=model.stat)
        assert SystemInfo().get_cpu_usage() == model.expect

    @staticmethod
    def test_get_cpu_stat_info(model: GetCpuStatInfo):
        assert SystemInfo().get_cpu_stat_info(model.info_list, model.logical) == model.expect

    @staticmethod
    def test_get_os_version_info(mocker: MockerFixture, model: GetOsVersionInfo):
        mocker.patch("builtins.open").return_value.__enter__.return_value.read.side_effect = model.read
        info = SystemInfo()
        info.get_os_version_info()
        assert info.OSVersion == model.expect

    @staticmethod
    def test_get_kernel_number(mocker: MockerFixture, model: GetKernelNumber):
        mocker.patch.object(SystemInfo, "get_kernel_version", return_value=model.version)
        info = SystemInfo()
        info.get_kernel_number()
        assert info.KernelVersion == model.expect

    @staticmethod
    def test_get_sys_electronic_tags(mocker: MockerFixture, model: GetSysElectronicTags):
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.return_value = model.sys
        info = SystemInfo()
        info.get_sys_electronic_tags()
        assert info.AssetTag == model.expect

    @staticmethod
    def test_get_host_name(mocker: MockerFixture, model: GetHostName):
        mocker.patch.object(socket, "gethostname", return_value=model.name)
        info = SystemInfo()
        info.get_host_name()
        assert info.HostName == model.expect

    @staticmethod
    def test_get_sys_uuid(mocker: MockerFixture, model: GetSysUuid):
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.return_value = model.uuid
        info = SystemInfo()
        info.get_sys_uuid()
        assert info.UUID == model.expect

    @staticmethod
    def test_get_mcu_temperature(mocker: MockerFixture, model: GetMcuTemperature):
        mocker.patch.object(Device, "get_attribute", return_value=model.board)
        info = SystemInfo()
        info.main_board = Device
        info.get_mcu_temperature()
        assert info.Temperature == model.expect

    @staticmethod
    def test_get_mcu_voltage(mocker: MockerFixture, model: GetMcuVoltage):
        mocker.patch.object(Device, "get_attribute", return_value=model.voltage)
        info = SystemInfo()
        info.main_board = Device
        info.get_mcu_voltage()
        assert info.Voltage == model.expect

    @staticmethod
    def test_get_power(mocker: MockerFixture, model: GetPower):
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.return_value = model.power
        mocker.patch.object(SystemSetting, "get_board_type").return_value = "Atlas A500 A2"
        mocker.patch.object(SystemUtils, "get_power_by_npu_smi").return_value = "2.0"
        info = SystemInfo()
        info.get_power()
        assert info.Power == model.expect

    @staticmethod
    def test_get_minid_temperature(mocker: MockerFixture, model: GetMinidTemperature):
        mocker.patch.object(DEVM, "get_device_list_by_module", return_value=model.npu)
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.return_value = model.tmp
        info = SystemInfo()
        info.get_minid_temperature()
        assert info.AiTemperature == model.expect

    @staticmethod
    def test_set_time(mocker: MockerFixture, model: SetTime):
        mocker.patch.object(SystemInfo, "set_os_time", return_value=model.set_os_time)
        assert SystemInfo().set_time({}, model.time) == model.expect

    @staticmethod
    def test_format_time(model: FormatTime):
        assert SystemInfo.format_time(model.time) == model.expect

    @staticmethod
    def test_set_os_time(mocker: MockerFixture, model: SetOSTime):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output", return_value=model.cmd)
        ret = SystemInfo.set_os_time("2023-02-27")
        assert ret == model.expect
