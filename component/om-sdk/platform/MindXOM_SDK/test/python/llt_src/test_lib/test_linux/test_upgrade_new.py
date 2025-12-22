from collections import namedtuple

import pytest
from pytest_mock import MockFixture

from bin.environ import Env
from common.file_utils import FileCheck
from common.file_utils import FileCreate
from common.file_utils import FileUtils
from common.utils.exec_cmd import ExecCmd
from common.utils.result_base import Result
from lib.Linux.systems.systems import SystemInfo
from lib.Linux.upgrade.schemas import ModuleState
from ut_utils.mock_utils import mock_cdll, mock_read_data

with mock_cdll():
    from lib.Linux.upgrade.upgrade_new import Upgrade, UpgradeError

PACKAGE_PATH = "lib.Linux.upgrade.upgrade_new"

UpgradeIsRunning = namedtuple("UpgradeIsRunning", "expect, state")
GetFirmwareActState = namedtuple("GetFirmwareActState", "expect, uboot, kernal, rootfs, minid")
GetAllInfo = namedtuple("GetAllInfo", "expect, state")
UdsUsernameAndIp = namedtuple("UdsUsernameAndIp", "expect, fd_ip")
GetFirmwareVersion = namedtuple("GetFirmwareVersion", "expect, mcu, uboot, os, minid")
GetProgress = namedtuple("GetProgress", "expect, module, process")
DrvUpgradeProcess = namedtuple("DrvUpgradeProcess", "expect, exist, file_size read_data")
VerifyInputParam = namedtuple("VerifyInputParam", "expect, request_data, check_path exception")
RemoveCreateDir = namedtuple("RemoveCreateDir", "expect delete create exception")
CmsVerify = namedtuple("CmsVerify", "expect check_path cms exception")
ExecShell = namedtuple("ExecShell", "expect check_script exec_script exception")
CreateFlagFile = namedtuple("CreateFlagFile", "expect exception")
GetModules = namedtuple("GetModules", "expect modules")
OmsdkUpgraded = namedtuple("OmsdkUpgraded", "expect state")
ChangeModuleState = namedtuple("ChangeModuleState", "expect exception")
AllowEffect = namedtuple("AllowEffect", "expect state")
ParseModuleMsg = namedtuple("ParseModuleMsg", "expect xml_parser start_m2 exception")


class TestUpgradeNew:
    use_cases = {
        "test_upgrade_is_running": {
            "running": (True, 1),
            "not running": (False, 0),
        },
        "test_get_all_info": {
            "new": ("New", 0),
            "Running": ("Running", 1),
            "Failed": ("Failed", 2),
            "Success": ("Success", 3),
        },
        "test_get_progress": {
            "not exists": (1, ModuleState(name="MCU", version="3.0.0"), 1),
            "normal": (0, ModuleState(name="MindXOM", version="3.0.0"), 0),
        },
        "test_get_drv_upgrade_process": {
            "file_not_exits": (0, False, 1024, "10"),
            "file_too_large": (0, True, 1024 * 1024 * 1042, "10"),
            "read_file_failed": (0, True, 1024, "0"),
            "get_process": (10, True, 1024, "10")

        },
        "test_verify_input_param": {
            "not_dict": ("1.zip", [], Result(result=False), UpgradeError),
            "invalid_ip_user": ("1.zip", {"_User": "", "_Xip": "192.168.1.111"}, Result(result=False), UpgradeError),
            "invalid_proc": ("1.zip", {"_User": "admin", "_Xip": "192.168.1.111",
                                       "TransferProtocol": "xxx"}, Result(result=False), UpgradeError),
            "not_zip": ("1.zip", {"_User": "admin", "_Xip": "192.168.1.111",
                                  "TransferProtocol": "https", "ImageURI": "1.txt"},
                        Result(result=False), UpgradeError),
            "invalid_path": ("1.zip", {"_User": "admin", "_Xip": "192.168.1.111",
                                       "TransferProtocol": "https", "ImageURI": "1.zip"},
                             Result(result=False), UpgradeError),
            "param_ok": ("1.zip", {"_User": "admin", "_Xip": "192.168.1.111",
                                   "TransferProtocol": "https", "ImageURI": "1.zip"}, Result(result=True), None),
        },
        "test_remove_and_create_dir": {
            "delete_dir_failed": (None, UpgradeError, None, UpgradeError),
            "create_dir_failed": (None, None, UpgradeError, UpgradeError),
            "success": (None, None, None, None),
        },
        "test_cms_verify": {
            "check_path_failed": (None, Result(False), False, UpgradeError),
            "verify_failed": (None, Result(True), False, UpgradeError),
            "success": (None, Result(True), True, None)
        },
        "test_exec_shell": {
            "script_check_failed": (None, Result(False), 1, UpgradeError),
            "exec_script_failed": (None, Result(True), 1, UpgradeError),
            "exec_success": (None, Result(True), 0, None)
        },
        "test_create_flag_file": {
            "create_failed": (None, UpgradeError),
            "create_success": (None, None)
        },
        "test_get_modules": {
            "none_module": ({}, {}),
            "you_module": ({"mcu": "module"}, {"mcu": "module"})
        },
        "test_omsdk_upgraded": {
            "no_upgrade": (False, False),
            "you_upgrade": (True, True),
        },
        "test_change_module_state": {
            "normal": (None, None),
            "exception": (None, Exception)
        },
        "test_allow_effect": {
            "not_allow": (False, False),
            "allow": (True, True)
        },
        "test_parse_module_msg": {
            "parse_success": ("MindXOM", ("MindXOM", "1.2"), False, None),
        }
    }

    @staticmethod
    def test_parse_module_msg(mocker: MockFixture, model: ParseModuleMsg):
        mocker.patch.object(Upgrade, "cms_verify")
        version_xml = namedtuple("version_xml", ["module", "version"])
        xml_parser = version_xml(*model.xml_parser)
        with mocker.patch("lib.Linux.upgrade.upgrade_new.VersionXmlManager",
                          side_effect=model.exception, return_value=xml_parser):
            mocker.patch.object(Env, "start_from_m2", return_value=model.start_m2)
            mocker.patch.object(Upgrade, "change_module_state")

            mocker.patch("lib.Linux.upgrade.upgrade_new.VersionXmlManager", return_value=xml_parser)
            assert Upgrade().parse_module_msg("version.xml") == model.expect

    @staticmethod
    def test_allow_effect(mocker: MockFixture, model: AllowEffect):
        Upgrade.modules = {"MindXOM": ModuleState(version="1.1", name="MindXOM", state=model.state)}
        assert Upgrade.allow_effect() == model.expect

    @staticmethod
    def test_change_module_state(mocker: MockFixture, model: ChangeModuleState):
        Upgrade.cur_module = ModuleState(version="1.1", name="MindXOM")
        with mocker.patch.object(SystemInfo, "get_all_info", side_effect=model.exception):
            assert Upgrade.change_module_state(process=10) == model.expect

    @staticmethod
    def test_omsdk_upgraded(mocker: MockFixture, model: OmsdkUpgraded):
        Upgrade.modules = {"MindXOM": ModuleState(version="1.1", name="MindXOM", state=model.state)}
        assert Upgrade.omsdk_upgraded() == model.expect

    @staticmethod
    def test_get_modules(mocker: MockFixture, model: GetModules):
        Upgrade.modules = model.modules
        assert Upgrade.get_modules() == model.expect

    @staticmethod
    def test_create_flag_file(mocker: MockFixture, model: CreateFlagFile):
        with mocker.patch("os.fdopen", side_effect=model.exception):
            if not model.exception:
                assert Upgrade().create_flag_file("flag") == model.expect
                return
            with pytest.raises(model.exception):
                Upgrade().create_flag_file("flag")

    @staticmethod
    def test_exec_shell(mocker: MockFixture, model: ExecShell):
        mocker.patch.object(FileUtils, "check_script_file_valid", return_value=model.check_script)
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=model.exec_script)
        if not model.exception:
            assert Upgrade().exec_shell("1.sh") == model.expect
            return
        with pytest.raises(model.exception):
            Upgrade().exec_shell("1.sh")

    @staticmethod
    def test_cms_verify(mocker: MockFixture, model: CmsVerify):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        mocker.patch("lib.Linux.upgrade.upgrade_new.verify_cms_file", return_value=model.cms)
        if not model.exception:
            assert Upgrade().cms_verify("1.tar.gz") == model.expect
            return
        with pytest.raises(model.exception):
            Upgrade().cms_verify("1.tar.gz")

    @staticmethod
    def test_remove_and_create_dir(mocker: MockFixture, model: RemoveCreateDir):
        mocker.patch.object(FileUtils, "delete_full_dir", side_effect=model.delete)
        mocker.patch.object(FileCreate, "create_dir", side_effect=model.create)
        if not model.exception:
            assert Upgrade().remove_and_create_dir("1.zip", 0o755) == model.expect
            return
        with pytest.raises(model.exception):
            Upgrade().remove_and_create_dir("1.zip", 0o755)

    @staticmethod
    def test_verify_input_param(mocker: MockFixture, model: VerifyInputParam):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        if not model.exception:
            assert Upgrade().verify_input_param(model.request_data) == model.expect
            return
        with pytest.raises(model.exception):
            Upgrade().verify_input_param(model.request_data)

    @staticmethod
    def test_get_drv_upgrade_process(mocker: MockFixture, model: DrvUpgradeProcess):
        Upgrade.cur_module = ModuleState(process=0)
        mocker.patch("os.path.exists", return_value=model.exist)
        mocker.patch("os.path.getsize", return_value=model.file_size)
        mock_read_data(mocker, read_data=model.read_data)
        assert Upgrade().get_drv_upgrade_process() == model.expect

    @staticmethod
    def test_upgrade_is_running(mocker: MockFixture, model: UpgradeIsRunning):
        Upgrade.upgrade_state = model.state
        assert Upgrade().is_running() == model.expect

    @staticmethod
    def test_get_all_info(mocker: MockFixture, model: GetAllInfo):
        mocker.patch.object(Upgrade, "get_progress")
        Upgrade.upgrade_proc_msg = "Running"
        Upgrade.upgrade_state = model.state
        upgrade = Upgrade()
        upgrade.get_all_info()
        assert upgrade.upgrade_proc_msg == model.expect

    @staticmethod
    def test_get_progress(mocker: MockFixture, model: GetProgress):
        Upgrade.cur_module = model.module
        mocker.patch.object(Upgrade, "get_drv_upgrade_process", return_value=model.process)
        assert Upgrade().get_progress() == model.expect
