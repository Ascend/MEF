from collections import namedtuple

import pytest
from pytest_mock import MockerFixture

from common.exception.biz_exception import BizException
from common.file_utils import FileCheck
from common.utils.app_common_method import AppCommonMethod
from common.utils.exec_cmd import ExecCmd
from conftest import TestBase


class TestGetJsonInfoObj:

    def __init__(self, obj_dict: dict, obj_list: list, obj_class: object):
        self.obj_dict = obj_dict
        self.obj_list = obj_list
        self.obj_class = obj_class


class TestSubGetJsonInfoObj:

    def __init__(self, name: str):
        self.name = name


class TestAppCommonMethod(TestBase):
    GetJsonErrorCase = namedtuple("GetJsonErrorCase", "excepted_status, excepted_message, result, default_status")
    ForceRemoveCase = namedtuple("ForceRemoveCase", "excepted, exists, islink, path_valid, isdir")
    GetDirectoryFreeSizeCase = namedtuple("GetDirectoryFreeSizeCase",
                                          "excepted, abs, exist_and_valid, exec_cmd, file_path")
    ExecCmdGetOutputCase = namedtuple("ExecCmdGetOutputCase", "excepted, cmd, code, out, error")
    ReplaceKvListCase = namedtuple("ReplaceKvListCase", "adict, copy, compare_key")
    ReplaceKvCase = namedtuple("ReplaceKvCase", "adict, k, v, key_type, compare_key")
    CheckIpv4FormatCase = namedtuple("CheckIpv4FormatCase", "excepted, ip_str")
    CheckNetworkPathCase = namedtuple("CheckNetworkPathCase", "excepted, is_link, path")
    CheckInputParmCase = namedtuple("CheckInputParmCase", "excepted, parm")
    GetKeyValuePairCase = namedtuple("GetKeyValuePairCase", "excepted, org_str, split_char")
    MakeAuthenticationString = namedtuple("MakeAuthenticationString", "expect, user, pwd")

    use_cases = {
        "test_get_json_error_by_array": {
            "result_dict": (200, "test", {"status": 200, "message": "test"}, 400),
            "result_list_int_not_exist": (404, "test", [404, "test"], 200),
            "result_list_int": (400, "test", [500, "test"], 200),
            "result_list_not_int": (200, ["500", "test"], ["500", "test"], 200),
            "result_list_len_1": (200, ["test"], ["test"], 200),
            "result_str": (200, "test", "test", 200),
        },
        "test_force_remove_file": {
            "exists_false": (True, False, False, False, False),
            "islink_true": (True, True, True, False, False),
            "check_input_path_valid": (False, True, False, False, False),
            "isdir_true": (True, True, False, True, True),
            "isdir_false": (True, True, False, True, False),
        },
        "test_check_ipv4_format": {
            "check_failed": (False, "51.0.0"),
            "startswith_127": (False, "127.0.0.1"),
            "all_zero": (False, "0.0.0.0"),
            "all_255": (False, "255.255.255.255"),
            "success": (True, "51.0.0.0"),
        },
        "test_check_input_parm": {
            "not_param": (True, ""),
            "param_not_str": (False, [1, 2]),
            "param_tow_point": (False, "..a"),
            "success": (True, "home.data"),
        },
        "test_get_key_value_pair": {
            "not_org": (["", ""], "", "a"),
            "not_split": (["", ""], "a:b", ""),
            "param_tow_point": (["", ""], "a:b", ";"),
            "success": (["a", "b"], "a:b", ":"),
        },
        "test_replace_kv_list": {
            "copy_is_list": ({1: 5}, [0, 1], 1),
            "copy_is_not_list": ({"test": 5}, {"test": 6}, "test")
        },
        "test_replace_kv": {
            "adict_is_list_equal": ([0, 5], 0, 1, None, None),
            "adict_is_list_key_is_list": ([[0, 1]], 0, 1, list, 0),
            "adict_is_list_key_is_dict": ([{"test": 5}], "test", 6, list, 0),
            "adict_is_not_list_equal": ({"test": 5}, "test", 6, None, None),
        },
        "test_make_authentication_string": {
            "normal": ("Basic bmFtZTpwd2Q=", "name", "pwd")
        },
    }

    def test_get_json_info(self):
        obj_dict = {"test": "anything"}
        obj_list = [1, 2, 3]
        ret = AppCommonMethod.get_json_info(
            TestGetJsonInfoObj(obj_dict, obj_list, TestSubGetJsonInfoObj("test_sub")))
        assert ret["obj_dict"] == {"test": "anything"}

    def test_get_project_absolute_path(self):
        ret = AppCommonMethod.get_project_absolute_path()
        assert ret is not None

    def test_get_json_error_by_array(self, model: GetJsonErrorCase):
        ret = AppCommonMethod.get_json_error_by_array(model.result, model.default_status)
        assert model.excepted_status == ret["status"]
        assert model.excepted_message == ret["message"]

    def test_force_remove_file(self, mocker: MockerFixture, model: ForceRemoveCase):
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch("os.path.islink", return_value=model.islink)
        mocker.patch.object(FileCheck, "check_input_path_valid", return_value=model.path_valid)
        mocker.patch("os.path.isdir")
        mocker.patch("os.remove")
        mocker.patch("shutil.rmtree")
        ret = AppCommonMethod.force_remove_file("/a/test")
        assert model.excepted == bool(ret)

    def test_replace_kv_list(self, model: ReplaceKvListCase):
        AppCommonMethod.replace_kv_list(model.adict, model.copy)
        assert model.adict[model.compare_key] == model.copy[model.compare_key]

    def test_replace_kv(self, model: ReplaceKvCase):
        AppCommonMethod.replace_kv(model.adict, model.k, model.v)
        if model.key_type == list:
            assert model.adict[model.compare_key][model.k] == model.v
        elif model.key_type == dict:
            for key in model.adict.keys:
                if key == model.compare_key:
                    assert key[model.k] == model.v
        else:
            assert model.adict[model.k] == model.v

    def test_check_ipv4_format(self, mocker: MockerFixture, model: CheckIpv4FormatCase):
        ret = AppCommonMethod.check_ipv4_format(model.ip_str)
        assert model.excepted == ret

    def test_hostname_check(self):
        ret = AppCommonMethod.hostname_check("abc-abc")
        assert ret

    def test_check_input_parm(self, model: CheckInputParmCase):
        ret = AppCommonMethod.check_input_parm(model.parm)
        assert model.excepted == ret

    def test_get_key_value_pair(self, model: GetKeyValuePairCase):
        ret = AppCommonMethod.get_key_value_pair(model.org_str, model.split_char)
        assert model.excepted == ret

    def test_make_authentication_string(self, model: MakeAuthenticationString):
        assert AppCommonMethod.make_authentication_string(model.user, model.pwd) == model.expect

    def test_make_authentication_string_exception(self):
        with pytest.raises(BizException):
            assert AppCommonMethod.make_authentication_string("", "")
