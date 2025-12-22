from collections import namedtuple
from configparser import ConfigParser
from typing import BinaryIO, IO
from unittest.mock import mock_open

import pytest
from pytest_mock import MockerFixture

from common.utils.result_base import Result

from common.exception.biz_exception import BizException
from common.file_utils import FileUtils, FilePermission, FileCopy, FileWriter
from common.file_utils import FileCheck


class TestFileUtils:
    FileSectionExceptionCase = namedtuple("FileSectionExceptionCase", "check_path, get_config_parser")
    FileModifySectionExceptionCase = namedtuple("FileModifySectionExceptionCase", "check_path, exists")
    ReadFileCase = namedtuple("ReadFileCase", "excepted, check_path, open")
    CopySectionSrcToDestCase = namedtuple("CopySectionSrcToDestCase", "excepted, check_path, check_section, "
                                                                      "get_option_list, get_option")
    CheckXmlFileValidCase = namedtuple("CheckXmlFileValidCase", "excepted, check_path")
    CheckRedfishMetadataXmlCase = namedtuple("CheckRedfishMetadataXmlCase",
                                             "excepted, path, exists, path_spec, whitelist, "
                                             "check_normal_file_path, is_link")
    _CheckNormalFilePath = namedtuple("_CheckNormalFilePath", "excepted, path")
    _CheckXmlIsSafe = namedtuple("_CheckXmlIsSafe", "excepted, getsize, open")
    SetPathOwnerGroupCase = namedtuple("SetPathOwnerGroupCase", "excepted, recursive, isfile")
    CopyFileCase = namedtuple("CopyFileCase", "excepted, isfile, check_path, check_input_path, mode, user, group,"
                                              " set_path_permission, set_path_owner_group")
    WriteAppendCase = namedtuple("WriteCase", "excepted, check_path")
    DeleteDirContentCase = namedtuple("DeleteDirContentCase", "expected, check_path")

    use_cases = {
        "test_get_option_list_exception": {
            "check_path_failed": (False, None),
            "not_has_section": (True, ConfigParser())
        },
        "test_get_option_exception": {
            "check_path_failed": (False, None),
            "not_has_section": (True, ConfigParser())
        },
        "test_modify_one_option_exception": {
            "check_path_failed": (False, None),
            "exists": (True, False),
        },
        "test_read_file": {
            "check_path_failed": ([], False, None),
            "normal": (["test"], True, "test"),
        },
        "test_check_xml_file_valid": {
            "check_path_failed": (False, False),
            "normal": (True, True)
        },
        "test_check_redfish_metadata_xml": {
            "path_null": (False, None, None, None, None, None, None),
            "path_not_str": (False, [1], None, None, None, None, None),
            "path_not_exists": (False, "/home/test", False, None, None, None, None),
            "path_len_wrong": (False, "/home/test" * 200, True, None, None, None, None),
            "not_whitelist": (False, "/home/test", True, True, False, None, None),
            "check_normal_file_path_failed": (False, "/home/test", True, True, True, False, None),
            "check_is_link": (False, "/home/test", True, True, True, True, False),
            "normal": (True, "/home/test", True, True, True, True, True)
        },
        "test_check_normal_file_path": {
            "not_match": (False, "#$%"),
            "two_point": (False, "../home"),
            "normal": (True, "a/home"),
        },
        "test_check_xml_is_safe": {
            "size_wrong": (False, 513 * 1024, None),
            "contain_key": (False, 1024, "<!DOCTYPEabc"),
            "normal": (True, 1024, "abc")
        },
        "test_set_path_owner_group": {
            "recursive_false": (True, False, None),
            "is_file": (True, True, True),
        },
        "test_copy_file": {
            "not_isfile": (False, False, None, None, None, None, None, None, None),
            "check_path_src_failed": (False, True, [False, None], None, None, None, None, None, None),
            "check_path_dst_failed": (False, True, [True, False], None, None, None, None, None, None),
            "check_input_path_failed": (False, True, [True, True], False, None, None, None, None, None),
            "set_path_permission_failed": (False, True, [True, True], True, "r", None, None, False, None),
            "user_not_null_group_null": (False, True, [True, True], True, "r", "r", None, True, None),
            "user_null_group_not_null": (False, True, [True, True], True, "r", None, "r", True, None),
            "set_path_owner_group_failed": (False, True, [True, True], True, "r", "r", "r", True, False),
            "normal": (True, True, [True, True], True, "r", "r", "r", True, True),
        },
        "test_write": {
            "check_failed": (False, False),
            "normal": (True, True),
        },
        "test_write_exception": {
            "exception": (False, True),
        },
        "test_append": {
            "check_failed": (False, False),
            "normal": (True, True),
        },
        "test_append_exception": {
            "exception": (False, True),
        },
        "test_delete_dir_content": {
            "normal": (True, Result(True))
        },
    }

    class UserInfo:
        def __init__(self, pw_uid, pw_gid):
            self.pw_uid = pw_uid
            self.pw_gid = pw_gid

    def test_get_config_parser_exception(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", return_value=False)
        with pytest.raises(BizException):
            mocker.patch("builtins.open")
            FileUtils.get_config_parser("/home/test")

    def test_get_config_parser_normal(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", return_value=True)
        mocker.patch.object(ConfigParser, "read")
        mocker.patch("builtins.open")
        ret = FileUtils.get_config_parser("/home/test")
        assert isinstance(ret, ConfigParser)

    def test_get_section_list(self, mocker: MockerFixture):
        parser = ConfigParser()
        parser.add_section("test")
        mocker.patch.object(FileUtils, "get_config_parser", return_value=parser)
        ret = FileUtils.get_section_list("/home/test")
        assert ret == ["test"]

    def test_get_option_list_exception(self, mocker: MockerFixture, model: FileSectionExceptionCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid",
                            return_value=model.check_path)
        mocker.patch.object(FileUtils, "get_config_parser", return_value=model.get_config_parser)
        with pytest.raises(BizException):
            FileUtils.get_option_list("/home/test", "test")

    def test_get_option_list_normal(self, mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=True)
        parser = ConfigParser()
        parser.read_dict({"test": {"test": "abc"}})
        mocker.patch.object(FileUtils, "get_config_parser", return_value=parser)
        ret = FileUtils.get_option_list("/home/test", "test")
        assert ret == {"test": "abc"}

    def test_get_option_exception(self, mocker: MockerFixture, model: FileSectionExceptionCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid",
                            return_value=model.check_path)
        mocker.patch.object(FileUtils, "get_config_parser", return_value=model.get_config_parser)
        with pytest.raises(BizException):
            FileUtils.get_option("/home/test", "test", "test")

    def test_get_option_normal(self, mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=True)
        parser = ConfigParser()
        parser.read_dict({"test": {"test": "abc"}})
        mocker.patch.object(FileUtils, "get_config_parser", return_value=parser)
        ret = FileUtils.get_option("/home/test", "test", "test")
        assert ret == "abc"

    def test_check_section_exist(self, mocker: MockerFixture):
        mocker.patch.object(FileUtils, "get_config_parser", return_value=ConfigParser())
        ret = FileUtils.check_section_exist("/home/test", "test")
        assert ret is False

    def test_add_one_section_exception(self, mocker: MockerFixture):
        mocker.patch.object(FileUtils, "check_section_exist", return_value=True)
        with pytest.raises(BizException):
            FileUtils.add_one_section("/home/test", "test", {"test": {"test": "abc"}})

    def test_add_one_section_normal(self, mocker: MockerFixture):
        mocker.patch.object(FileUtils, "check_section_exist", return_value=False)
        mocker.patch.object(FileUtils, "operate_file")
        ret = FileUtils.add_one_section("/home/test", "test", {"test": "abc"})
        assert ret is None

    def test_modify_one_option_exception(self, mocker: MockerFixture, model: FileModifySectionExceptionCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid",
                            return_value=model.check_path)
        mocker.patch.object(FileUtils, "check_section_exist", return_value=model.exists)
        with pytest.raises(BizException):
            FileUtils.modify_one_option("/home/test", "test", "test", "abc")

    def test_modify_one_option_normal(self, mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=True)
        mocker.patch.object(FileUtils, "check_section_exist", return_value=True)
        parser = ConfigParser()
        parser.read_dict({"test": {"test": "abc"}})
        mocker.patch.object(FileUtils, "get_config_parser", return_value=parser)
        mocker.patch.object(FileUtils, "operate_file")
        ret = FileUtils.modify_one_option("/home/test", "test", "test", "abc")
        assert ret is None

    def test_operate_file_exception(self, mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_is_link", return_value=False)
        with pytest.raises(BizException):
            FileUtils.operate_file(ConfigParser(), "/home/test", "r")

    def test_operate_file_normal(self, mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_is_link", return_value=True)
        mocker.patch("builtins.open")
        mocker.patch.object(ConfigParser, "write")
        ret = FileUtils.operate_file(ConfigParser(), "/home/test", "r")
        assert ret is None

    def test_read_file(self, mocker: MockerFixture, model: ReadFileCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid",
                            return_value=model.check_path)
        mocker.patch("builtins.open", mock_open(read_data=model.open))
        ret = FileUtils.read_file("/home/test", "r")
        assert model.excepted == ret

    def test_check_xml_file_valid(self, mocker: MockerFixture, model: CheckXmlFileValidCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid",
                            return_value=model.check_path)
        mocker.patch.object(FileCheck, "_check_xml_is_safe", return_value=True)
        ret = FileCheck.check_xml_file_valid("/home/test.xml")
        assert model.excepted == ret

    def test_check_normal_file_path(self, model: _CheckNormalFilePath):
        ret = FileCheck._check_normal_file_path(model.path)
        assert model.excepted == bool(ret)

    def test_check_xml_is_safe(self, mocker: MockerFixture, model: _CheckXmlIsSafe):
        mocker.patch("os.path.getsize", return_value=model.getsize)
        mocker.patch("builtins.open", mock_open(read_data=model.open))
        ret = FileCheck._check_xml_is_safe("/home/test")
        assert model.excepted == bool(ret)

    def test_check_xml_is_safe_exception(self, mocker: MockerFixture):
        mocker.patch("os.path.getsize", return_value=1024)
        mocker.patch("builtins.open", side_effect=Exception())
        ret = FileCheck._check_xml_is_safe("/home/test")
        assert bool(ret) is False

    def test_set_path_owner_group(self, mocker: MockerFixture, model: SetPathOwnerGroupCase):
        user = TestFileUtils.UserInfo("1", "1")
        mocker.patch("pwd.getpwnam", return_value=user)
        mocker.patch("os.path.isfile", return_value=model.isfile)
        mocker.patch("os.lchown")
        mocker.patch("os.walk", ["root", "test", "a.log"])
        ret = FilePermission.set_path_owner_group("/home/test", "root")
        assert model.excepted == bool(ret)

    def test_set_path_owner_group_exception(self, mocker: MockerFixture):
        mocker.patch("pwd.getpwnam", side_effect=Exception())
        ret = FilePermission.set_path_owner_group("/home/test", "root")
        assert bool(ret) is False

    def test_copy_file(self, mocker: MockerFixture, model: CopyFileCase):
        mocker.patch("os.path.isfile", return_value=model.isfile)
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", side_effect=model.check_path)
        mocker.patch.object(FileCheck, "check_input_path_valid", return_value=model.check_input_path)
        mocker.patch("shutil.copyfile")
        mocker.patch.object(FilePermission, "set_path_permission",
                            return_value=model.set_path_permission)
        mocker.patch.object(FilePermission, "set_path_owner_group",
                            return_value=model.set_path_owner_group)
        ret = FileCopy.copy_file("/home/test1", "/home/test2", model.mode, model.user, model.group)
        assert model.excepted == bool(ret)

    def test_copy_file_exception(self, mocker: MockerFixture):
        mocker.patch("os.path.isfile", return_value=True)
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", side_effect=[True, True])
        mocker.patch.object(FileCheck, "check_input_path_valid", return_value=True)
        mocker.patch("shutil.copyfile", side_effect=Exception())
        ret = FileCopy.copy_file("/home/test1", "/home/test2")
        assert "copy file error" in ret.error

    def test_write(self, mocker: MockerFixture, model: WriteAppendCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        mocker.patch("builtins.open")
        mocker.patch.object(BinaryIO, "write")
        ret = FileWriter("/home/test").write("test")
        assert model.excepted == bool(ret)

    def test_write_exception(self, mocker: MockerFixture, model: WriteAppendCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        mocker.patch("builtins.open", side_effect=Exception())
        ret = FileWriter("/home/test").write("test")
        assert model.excepted == bool(ret)

    def test_append(self, mocker: MockerFixture, model: WriteAppendCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        mocker.patch("builtins.open")
        mocker.patch.object(IO, "writelines")
        ret = FileWriter("/home/test").append(["test"])
        assert model.excepted == bool(ret)

    def test_append_exception(self, mocker: MockerFixture, model: WriteAppendCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        mocker.patch("builtins.open", side_effect=Exception())
        ret = FileWriter("/home/test").append(["test"])
        assert model.excepted == bool(ret)
