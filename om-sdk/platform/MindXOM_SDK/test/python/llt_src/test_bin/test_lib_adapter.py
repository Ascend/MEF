#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import threading
from collections import namedtuple

from pytest_mock import MockerFixture

from bin.lib_adapter import LibAdapter
from common.file_utils import FileCheck
from common.restfull_socket_model import RestFullSocketModel


class TestLibAdapter:

    @staticmethod
    def test_init_resources(mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid").return_value = True
        mocker.patch("builtins.open")
        LibAdapter.init_resources()

    @staticmethod
    def test_init_resource_lock():
        class_path = {
            "SecurityService": {
                "action": {
                    "SecurityService_all": {
                        "hasList": "false",
                    }
                },
                "description": "Get SecurityService Info",
                "enabled": "true",
                "intervalTime": 3600,
                "runTimes": 0,
                "minIntervalTime": 1,
                "maxIntervalTime": 86400,
            },
        }
        LibAdapter.init_resource_lock(class_path)

    @staticmethod
    def test_lib_socket_call_function(mocker: MockerFixture):
        def get_request_body(request_type):
            return {
                "method": "lib_restful_interface",
                "model_name": "TestCase",
                "request_type": request_type,
                "request_data": None,
                "need_list": False,
                "item1": "UT",
            }

        ibma_class_json = {
            "TestCase": {
                "class": "lib.Linux.systems.devm.Module",
                "keys": "",
            }
        }
        mocker.patch.object(RestFullSocketModel, "check_socket_model").return_value = True
        mocker.patch.object(LibAdapter, "iBMAClassPath", ibma_class_json)
        for req_type in "GET", "POST", "OTHER":
            LibAdapter.lib_socket_call_function(get_request_body(req_type))

    @staticmethod
    def test_lib_timer_interface(mocker: MockerFixture):
        module = "TestCase"
        ibma_class_json = {
            module: {
                "class": "lib.Linux.systems.devm.Module",
                "keys": "",
            }
        }
        mocker.patch.object(LibAdapter, "iBMAClassPath", ibma_class_json)
        resource = {module: threading.Lock()}
        mocker.patch.object(LibAdapter, "ResourceLock", resource)
        LibAdapter.lib_timer_interface(module, "func", item1="test")

    @staticmethod
    def test_get_resource_info(mocker: MockerFixture):
        module = "TestCase"
        ibma_class_json = {
            module: {
                "path": "lib.Linux.systems.devm.Module",
                "keys": "",
            }
        }
        mocker.patch.object(LibAdapter, "iBMAClassPath", ibma_class_json)
        resource = {module: threading.Lock()}
        mocker.patch.object(LibAdapter, "ResourceLock", resource)
        resource = {f"{module}_test_None_None_None": {"func": "6666"}}
        mocker.patch.object(LibAdapter, "iBMAResources", resource)
        LibAdapter.get_resource_info(module, "func", item1="test")

    @staticmethod
    def test_get_ibma_class_path(mocker: MockerFixture):
        module = "TestCase"
        ibma_class_path = {
            module: {
                "path": "lib.Linux.systems.devm.Module",
                "keys": "",
            }
        }
        mocker.patch.object(LibAdapter, "iBMAClassPath", ibma_class_path)
        ret = LibAdapter.get_ibma_class_path().get(module)
        assert ret == ibma_class_path.get(module)

    @staticmethod
    def test_get_ibma_resources_value(mocker: MockerFixture):
        resource = {"Module_None_None_None_None": {"test": "6666"}}
        mocker.patch.object(LibAdapter, "iBMAResources", resource)
        ret = LibAdapter.get_ibma_resources_value("Module,test,None,None,None")
        assert ret == "6666"

    @staticmethod
    def test_call_function():
        def func(*args):
            return [arg for arg in args]

        Arg = namedtuple("Arg", "item1, item2, item3, item4")
        args = (
            Arg(item1=None, item2=2, item3=3, item4=4),
            Arg(item1=1, item2=None, item3=3, item4=4),
            Arg(item1=1, item2=2, item3=None, item4=4),
            Arg(item1=1, item2=2, item3=3, item4=None),
            Arg(item1=1, item2=2, item3=3, item4=4),
        )
        ret_dict = {0: [], 1: [1], 2: [1, 2], 3: [1, 2, 3]}
        for num, arg in enumerate(args):
            ret = LibAdapter.call_function(func, arg)
            assert ret == ret_dict.get(num, [1, 2, 3, 4])

    @staticmethod
    def test_call_function_with_request_data():

        def func(*args):
            return [arg for arg in args]

        Arg = namedtuple("Arg", "request_data, item1, item2, item3, item4")
        args = (
            Arg(request_data="6", item1=None, item2=2, item3=3, item4=4),
            Arg(request_data="66", item1=1, item2=None, item3=3, item4=4),
            Arg(request_data="666", item1=1, item2=2, item3=None, item4=4),
            Arg(request_data="6666", item1=1, item2=2, item3=3, item4=None),
            Arg(request_data="66666", item1=1, item2=2, item3=3, item4=4),
        )
        ret_dict = {
            0: [args[0].request_data],
            1: [args[1].request_data, 1],
            2: [args[2].request_data, 1, 2],
            3: [args[3].request_data, 1, 2, 3]
        }
        for num, arg in enumerate(args):
            ret = LibAdapter.call_function_with_request_data(func, arg)
            assert ret == ret_dict.get(num, [args[4].request_data, 1, 2, 3, 4])

    @staticmethod
    def test_delete_class_object_by_items(mocker: MockerFixture):
        module = "TestCase"
        path = "class"
        ibma_class_path = {
            module: {
                path: "lib.Linux.systems.devm.Module",
                "keys": "",
            }
        }
        mocker.patch.object(LibAdapter, "iBMAClassPath", ibma_class_path)
        ibma_class_json = {
            ibma_class_path.get(module).get(path): {
                f"{module}_1_2_3_4": 1,
            },
        }
        mocker.patch.object(LibAdapter, "iBMAClassObjs", ibma_class_json)
        LibAdapter.delete_class_object_by_items(module, 1, 2, 3, 4)
        assert LibAdapter.iBMAClassObjs.get(ibma_class_path.get(module).get(path)).get(f"{module}_1_2_3_4") is None

    @staticmethod
    def test_set_resource(mocker: MockerFixture):
        resource = {"Module": threading.Lock()}
        mocker.patch.object(LibAdapter, "ResourceLock", resource)
        for key in "all", "other":
            LibAdapter.set_resource("UT", "Module", key=key)

    @staticmethod
    def test_generate_event(mocker: MockerFixture):
        Param = namedtuple(
            "Param",
            "model_name, old_resource, resource, all_resource, old_is_all, p_path, c_path, item1"
        )
        module = "TestCase"
        params = (
            Param(
                model_name=module, old_resource=["1", "3"], resource=["1", "2"], all_resource=["1", "2", "3"],
                old_is_all=False, p_path=None, c_path="/tmp/test;1", item1=1
            ),
            Param(
                model_name=module, old_resource={"1": 1}, resource={"1": 1, "2": 2}, all_resource={"1": 1},
                old_is_all=True, p_path=None, c_path=None, item1=1
            ),
        )
        resource = {module: threading.Lock()}
        mocker.patch.object(LibAdapter, "ResourceLock", resource)
        for param in params:
            LibAdapter.generate_event(
                param.model_name, param.old_resource, param.resource, param.all_resource, param.old_is_all
            )

    @staticmethod
    def test_check_list():
        Param = namedtuple("Param", "old, new")
        params = (
            Param(old=[1, 2, 3], new=[2, 3, 4]),
            Param(old=[[1], [2], [3]], new=[[2], [3], [4]]),
        )
        ret_value = {
            0: [1, [2, 3, 4]],
            1: [1, [[2], [3], [4]]],
        }
        for num, param in enumerate(params):
            ret = LibAdapter.check_list(param.old, param.new)
            assert ret == ret_value.get(num)

    @staticmethod
    def test_replace_change_attributes():
        all_resource = {}
        msg = {1: 1, 2: 2}
        assert LibAdapter.replace_change_attributes(all_resource, msg) == msg

    @staticmethod
    def test_set_ibma_timers(mocker: MockerFixture):
        key = "date"
        param = {key: 10086}
        mocker.patch.object(LibAdapter, "ibma_timers", param)
        LibAdapter.set_ibma_timers(param)
        assert LibAdapter.ibma_timers.get(key) == param.get(key)

    @staticmethod
    def test_get_ibma_timers_cfg(mocker: MockerFixture):
        key = "date"
        param = {key: 10086}
        mocker.patch.object(LibAdapter, "ibma_timers_cfg", param)
        assert LibAdapter.get_ibma_timers_cfg() == param

    @staticmethod
    def test_set_ibma_timers_cfg(mocker: MockerFixture):
        key = "date"
        param = {key: 10086}
        mocker.patch.object(LibAdapter, "ibma_timers_cfg", param)
        LibAdapter.set_ibma_timers_cfg(param)
        assert LibAdapter.ibma_timers_cfg.get(key) == param.get(key)
