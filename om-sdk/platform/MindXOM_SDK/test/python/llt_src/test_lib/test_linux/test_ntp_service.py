from collections import namedtuple

from common.file_utils import FileCheck
from common.utils.exec_cmd import ExecCmd
from pytest_mock import MockerFixture

from lib.Linux.systems.ntp import NTPService
from common.common_methods import CommonMethods

NtpServersCmd = namedtuple("NtpServersCmd", "expect, cmd")
GetNtpServers = namedtuple("GetNtpServers", "expect, servers")
GetNtpInfo = namedtuple("GetNtpInfo", "expect, path_valid, read")
PatchRequest = namedtuple("PatchRequest", "expect, lock, check, client, remote_ck, update, stop addr")
StopNtpServer = namedtuple("StopNtpServer", "expect, cmd")
NtpServerFormatCheck = namedtuple("NtpServerFormatCheck", "expect, server")
CheckRequestDict = namedtuple("CheckRequestDict", "expect, request")
RemoteServerCheck = namedtuple("RemoteServerCheck", "expect, format_ck, server, server_bak")
GetLocalServerIp = namedtuple("GetLocalServerIp", "expect, listen")
UpdateNtpServer = namedtuple("UpdateNtpServer", "expect, format_ck, sign")


class TestNtpService:
    use_cases = {
        "test_ntp_servers_cmd": {
            "normal": ([], [1, "result\nresult\n"])
        },
        "test_get_ntp_servers": {
            "null bak": (("remote", None), ["remote"]),
            "normal": (("remote", "bak"), ["remote", "bak"])
        },
        "test_get_ntp_info": {
            "invalid path": (False, False, None),
            "read err": (False, True, Exception),
            "normal": (True, True, ["{'NTPClientEnable': True}"])
        },
        "test_patch_request": {
            "busy": ([CommonMethods.ERROR, "NTP modify is busy"], True, None, None, None, None, None, "51.38.64.2"),
            "invalid params": (
                [CommonMethods.ERROR, "invalid params"], False, [CommonMethods.ERROR, "invalid params"], None, None,
                None, None, "51.38.64.2"
            ),
            "invalid remote": (
                [CommonMethods.ERROR, "invalid remote"], False, [CommonMethods.OK, ""], True,
                [CommonMethods.ERROR, "invalid remote"], None, None, "51.38.64.2"
            ),
            "set success": (
                [CommonMethods.OK, "NTP set success"], False, [CommonMethods.OK, ""], True, [CommonMethods.OK, ""],
                [0, ""], None, "51.38.64.2"
            ),
            "update err": (
                [CommonMethods.ERROR, "update err"], False, [CommonMethods.OK, ""], True, [CommonMethods.OK, ""],
                [1, "update err"], None, "51.38.64.2"
            ),
            "set success-1": (
                [CommonMethods.OK, "NTP set success"], False, [CommonMethods.OK, ""], False,
                None, None, [0, ""], "51.38.64.2"
            ),
            "stop failed": (
                [CommonMethods.ERROR, "stop failed"], False, [CommonMethods.OK, ""], False, None, None,
                [1, "stop failed"], "51.38.64.2"
            )
        },
        "test_stop_ntp_server": {
            "stop failed": ([1, "Failed to stop ntp service"], 1),
            "stop success": ([0, "stop shell ok"], 0),
        },
        "test_ntp_server_format_check": {
            "null server": ([1, "Ntpserver is none"], None),
            "valid server": ([0, " NTP format is ok"], "10.10.10.10"),
            "invalid server": ([1, "ntpserver is illegal"], "")
        },
        "test_check_request_dict": {
            "NTPServerEnable wrong format": (
                [CommonMethods.ERROR, "NTPServerEnable wrong format"], {"ServerEnabled": None}
            ),
            "ntp_client_enable wrong format": (
                [CommonMethods.ERROR, "ntp_client_enable wrong format"], {"ClientEnabled": None}
            ),
            "Ntp target is not Client": (
                [CommonMethods.ERROR, "Ntp target is not Client"],
                {"ServerEnabled": True, "ClientEnabled": True, "Target": "any"}
            ),
            "no support NTPServerEnable config": (
                [CommonMethods.ERROR, "no support NTPServerEnable config"],
                {"ServerEnabled": True, "ClientEnabled": True}
            ),
            "NTPRemoteServer must be configured": (
                [CommonMethods.ERROR, "no support NTPServerEnable config"],
                {"ServerEnabled": True, "ClientEnabled": True, "Target": "Client"}
            ),
            "Servers ip wrong format": (
                [CommonMethods.ERROR, "Servers ip wrong format"],
                {"ServerEnabled": False, "ClientEnabled": True, "Target": "Client", "NTPRemoteServers": "0.0.0.0"}
            ),
            "check request dict success.": (
                [CommonMethods.OK, "check request dict success."],
                {"ServerEnabled": False, "ClientEnabled": True, "Target": "Client", "NTPRemoteServers": "10.0.0.0"}
            )
        },
        "test_remote_server_check": {
            "invalid server": (
                [CommonMethods.ERROR, "NTPRemoteServers :ntp remote server is illegal"], [[1, ]], None, None
            ),
            "invalid bak server": (
                [CommonMethods.ERROR, "NTPRemoteServersbak: ntp remote back server is illegal"], [[0, ], [1, ]], None,
                True
            ),
            "invalid remote": (
                [CommonMethods.ERROR, "The primary remote NTP address is the same as the secondary address."],
                [[0, ], [0, ]], "server", "server"
            ),
            "ok": ([CommonMethods.OK, "server;remote"], [[0, ], [0, ]], "server", "remote"),
            "ok-1": ([CommonMethods.OK, "server"], [[0, ], [0, ]], "server", ""),
        },
        "test_update_ntp_server": {
            "invalid local server": ([1, "NTPLocalServers : [local] is illegal"], [1, ], 0),
            "start ntp failed": ([1, "Failed to start ntp service"], [0, ], 1),
            "start ntp ok": ([0, "start ntp service ok"], [0, ], 0),
        }
    }

    @staticmethod
    def test_ntp_servers_cmd(mocker: MockerFixture, model: NtpServersCmd):
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol", return_value=model.cmd)
        assert NTPService.get_ntp_config() == model.expect

    @staticmethod
    def test_get_ntp_servers(mocker: MockerFixture, model: GetNtpServers):
        mocker.patch.object(NTPService, "get_ntp_config", return_value=model.servers)
        service = NTPService()
        service.get_ntp_servers()
        assert (service.NTPRemoteServers, service.NTPRemoteServersbak) == model.expect

    @staticmethod
    def test_get_ntp_info(mocker: MockerFixture, model: GetNtpInfo):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.path_valid)
        mocker.patch.object(FileCheck, "check_is_link_exception")
        mocker.patch("builtins.open").return_value.__enter__.return_value.read.side_effect = model.read
        assert NTPService().get_ntp_info() == model.expect

    @staticmethod
    def test_patch_request(mocker: MockerFixture, model: PatchRequest):
        mocker.patch.object(NTPService, "NTP_LOCK").locked.return_value = model.lock
        mocker.patch.object(NTPService, "check_request_dict", return_value=model.check)
        mocker.patch.object(NTPService, "remote_server_check", return_value=model.remote_ck)
        mocker.patch.object(NTPService, "update_ntp_server", return_value=model.update)
        mocker.patch("time.sleep")
        mocker.patch.object(NTPService, "stop_ntp_server", return_value=model.stop)
        mocker.patch.object(NTPService, "get_local_server_ip", return_value=model.stop)
        service = NTPService()
        service.ntp_client_enable = model.client
        assert service.patch_request({}) == model.expect

    @staticmethod
    def test_stop_ntp_server(mocker: MockerFixture, model: StopNtpServer):
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=model.cmd)
        assert NTPService().stop_ntp_server() == model.expect

    @staticmethod
    def test_ntp_server_format_check(model: NtpServerFormatCheck):
        assert NTPService().ntpserver_format_check(model.server) == model.expect

    @staticmethod
    def test_check_request_dict(model: CheckRequestDict):
        assert NTPService().check_request_dict(model.request) == model.expect

    @staticmethod
    def test_remote_server_check(mocker: MockerFixture, model: RemoteServerCheck):
        mocker.patch.object(NTPService, "ntpserver_format_check", side_effect=model.format_ck)
        assert NTPService().remote_server_check(model.server, model.server_bak) == model.expect

    @staticmethod
    def test_update_ntp_server(mocker: MockerFixture, model: UpdateNtpServer):
        mocker.patch.object(NTPService, "ntpserver_format_check", return_value=model.format_ck)
        mocker.patch("time.sleep")
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=model.sign)
        assert NTPService().update_ntp_server("", "local") == model.expect
