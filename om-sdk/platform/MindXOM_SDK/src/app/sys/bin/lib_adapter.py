#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

"""
功    能：插件适配，用于调用底层实现，以及反馈结果
修改记录：2016-10-31 创建
"""
import importlib
import json
import operator
import os
import threading
from collections import namedtuple

from bin.extend_interfaces import EXTEND_CLASS_MAP
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.restfull_socket_model import RestFullSocketModel as RestFullSocketModel
from common.utils.app_common_method import AppCommonMethod
from common.common_methods import CommonMethods

# 资源变动动作定义
RESOURCE_UPDATED = "ResourceUpdated"
RESOURCE_ADDED = "ResourceAdded"
RESOURCE_REMOVE = "ResourceRemoved"

FuncParams = namedtuple('FuncParams', ['request_data', 'item1', 'item2', 'item3', 'item4'])


class LibAdapter(object):
    """
    功能描述：插件适配
    接口：NA
    修改记录：2016-10-31 创建
    """

    # 定义互斥锁
    mutex = threading.Lock()

    # 用于存放临时数据信息
    iBMAResources = {}

    # 用于存放资源锁
    ResourceLock = {}

    # 用于存放类路径
    iBMAClassPath = {}

    iBMAClassObjs = {}

    # 事件上报资源列表，用于控制事件上报动作是否被处理
    eventResourcesDict = {
        RESOURCE_ADDED: {},
        RESOURCE_UPDATED: {},
        RESOURCE_REMOVE: {}
    }
    eventResourcesLock = threading.Lock()

    # 用于存放定时器
    ibma_timers = {}

    ibma_timers_cfg = {}

    @staticmethod
    def init_class_resources(classmap_path: str) -> dict:
        cfg_file_path = os.path.join(AppCommonMethod.get_project_absolute_path(), "config", classmap_path)

        ret = FileCheck.check_path_is_exist_and_valid(cfg_file_path)
        if not ret:
            run_log.error(f"%s invalid: %s", cfg_file_path, ret.error)
            return {}

        # 获取配置文件内容
        try:
            with open(cfg_file_path, "r") as file:
                return json.load(file)
        except Exception as err:
            run_log.error("read %s failed: %s", classmap_path, err)
            return {}

    @staticmethod
    def init_resources():
        """
        功能描述：初始化资源
        参数：无
        返回值：是否成功
        异常描述：NA
        修改记录：2016-11-23 创建
        """
        try:
            LibAdapter.mutex.acquire()
            # `not` on empty dict should return True
            if not LibAdapter.iBMAResources:
                LibAdapter.iBMAResources = {}
            else:
                return True

            # `not` on empty dict should return True
            if not LibAdapter.iBMAClassPath:
                LibAdapter.iBMAClassPath = LibAdapter.init_class_resources("ClassMap.json")
                # 初始化扩展的Classmap
                if EXTEND_CLASS_MAP:
                    LibAdapter.iBMAClassPath.update(LibAdapter.init_class_resources(EXTEND_CLASS_MAP))

            return True
        finally:
            LibAdapter.mutex.release()

    @staticmethod
    def init_resource_lock(class_path):
        """
        功能描述：初始化资源互斥锁
        参数：classPath 资源信息
        返回值：是否成功
        异常描述：NA
        修改记录：2016-12-06 创建
                2017-01-03 添加对子对象的资源锁初始化，防止出现不缓存的情况
        """
        # 初始化各自的资源锁，用于互斥操作
        for key in class_path.keys():
            is_local = class_path[key].get("isLocal", False)
            if is_local:
                # 本地的获取方式不需要加锁
                continue

            # 将根节点添加到锁列表中
            LibAdapter.ResourceLock[key] = threading.Lock()

            # 遍历 action 列表初始化互斥锁
            actions = class_path[key].get("action", None)
            if actions is None or not isinstance(actions, dict):
                continue

            for action_key in actions.keys():
                index = action_key.find("_")
                if index < 1:
                    continue

                action_key = action_key[0: index]
                if action_key not in LibAdapter.ResourceLock:
                    LibAdapter.ResourceLock[action_key] = threading.Lock()

    @staticmethod
    def lib_rest_full_model_process(json_data):
        from bin import monitor
        # 获取 Socket 模型
        socket_model = RestFullSocketModel.get_socket_model(json_data)
        if socket_model:
            if socket_model.method == "lib_restful_interface":
                return LibAdapter.lib_restful_interface(
                    socket_model.model_name, socket_model.request_type, socket_model.request_data,
                    socket_model.need_list, socket_model.item1, socket_model.item2,
                    socket_model.item3, socket_model.item4)
            elif socket_model.method == "start_timer":
                # 初始化定时器
                return monitor.Monitor.start_timer()
        return CommonMethods.object_to_json(CommonMethods.ERROR, "Request method is invalid.")

    @staticmethod
    def lib_socket_call_function(request_data):
        """
        功能描述：Socket 服务端的回调函数，用于接收并处理数据
        参数：request_data 请求的数据
        返回值：调用底层接口获取数据并返回
        异常描述：NA
        修改记录：2017-01-02 创建
        """

        # 首先将传入的参数转成 json
        ret = CommonMethods.check_json_data(request_data)
        if ret[0] != 0 or request_data is None:
            return CommonMethods.object_to_json(CommonMethods.ERROR, "Request data is invalid.")

        if ret[1] is None:
            return CommonMethods.object_to_json(CommonMethods.ERROR, "Request data is null.")

        result = RestFullSocketModel.check_socket_model(ret[1])
        if result:
            return LibAdapter.lib_rest_full_model_process(ret[1])

        return CommonMethods.object_to_json(CommonMethods.ERROR, "Request model is invalid.")

    @staticmethod
    def lib_restful_interface(model_name, request_type,
                              request_data, need_list=False,
                              item1=None, item2=None, item3=None, item4=None):
        """
        功能描述：底层接口的上层调用接口
        参数：modelName 模块名称，用于加载具体的对象
        requestType 请求类型，包括 POST，PATCH
        request_data 请求的属性
        needList 是否是一个列表，True 或 False
        item1 第一级资源
        item2 第二级资源
        item3 第三级资源
        item4 第四级资源
        返回值：无
        异常描述：NA
        修改记录：2016-11-07 创建
                2016-11-15 增加 request_data 参数，用于POST、PATCH数据传输
                2016-12-13 将底层发送命令调整到 lib_request_interface
                函数中，以便支持内部的 Start、Stop 指令
        """
        if request_type is None:
            return CommonMethods.object_to_json(CommonMethods.ERROR, "Request type can not be none.")

        if model_name is None or model_name not in LibAdapter.iBMAClassPath:
            return CommonMethods.object_to_json(CommonMethods.NOT_EXIST, "Model name is not found.")

        if request_type == "GET":
            if not isinstance(need_list, bool):
                run_log.error("%s needList value is error: %s." % (model_name, str(need_list)))
                need_list = False
            return LibAdapter.get_resource_info(model_name, "all", need_list, True, item1, item2, item3, item4)

        if request_type != "POST" and request_type != "PATCH" and request_type != "DELETE":
            # 不支持的类型
            return CommonMethods.object_to_json(CommonMethods.ERROR,
                                                "Request type:" + str(request_type) + "is not support.")

        # 调用底层接口进行操作
        return LibAdapter.lib_request_interface(model_name, request_type,
                                                request_data, need_list, item1,
                                                item2, item3, item4)

    @staticmethod
    def lib_request_interface(model_name, request_type, request_data,
                              need_list=False, item1=None, item2=None,
                              item3=None, item4=None):
        """
        功能描述：底层接口的上层调用接口
        参数：modelName 模块名称，用于加载具体的对象
        requestType 请求类型，包括 POST，PATCH，START，STOP
        request_data 请求的属性
        needList 是否是一个列表，True 或 False
        item1 第一级资源
        item2 第二级资源
        item3 第三级资源
        item4 第四级资源
        返回值：无
        异常描述：NA
        修改记录：2016-12-13 创建
                2016-11-15 增加 request_data 参数，用于POST、PATCH数据传输
                2016-12-23 判断传入的 requestData参数是否满足要求
        """
        if request_type is None:
            return CommonMethods.object_to_json(CommonMethods.ERROR, "Unsupported request type.")

        if model_name is None or model_name not in LibAdapter.iBMAClassPath:
            return CommonMethods.object_to_json(CommonMethods.NOT_EXIST, "Model name is not found.")

        func_name_dict = {
            "POST": "post_request",
            "PATCH": "patch_request",
            "DELETE": "delete_request",
            "START": "start_request",
            "STOP": "stop_request"
        }
        func_name = func_name_dict.get(request_type, None)
        # 不支持的类型
        if func_name is None:
            return CommonMethods.object_to_json(CommonMethods.ERROR,
                                                "Request type:" + str(request_type) + " is not support.")

        run_log.info(f"{model_name} call {request_type} function start.")

        # 判断传入的参数是否为 dict 类型，是否存在重复的键值，反馈错误信息
        ret = CommonMethods.check_json_data(request_data)
        if ret[0] != 0:
            return AppCommonMethod.get_json_error_by_array(ret)
        # 将数据设置为转换后的字典
        request_data = ret[1]
        try:
            # 获取底层接口函数
            tmp_ret = LibAdapter.get_function(LibAdapter.iBMAClassPath[model_name]["class"], func_name)
        except Exception as ex:
            run_log.error(f"{model_name} call {request_type} function failed, get function failed.")
            return CommonMethods.object_to_json(CommonMethods.INTERNAL_ERROR, str(ex))

        if not isinstance(tmp_ret, list):
            return tmp_ret

        class_obj = tmp_ret[0]
        func_params = FuncParams(request_data, item1, item2, item3, item4)
        try:
            tmp_ret = LibAdapter.call_function_with_request_data(tmp_ret[1], func_params)
        except Exception as ex:
            run_log.error(f"{model_name} call {request_type} function failed, call function failed.")
            return CommonMethods.object_to_json(CommonMethods.INTERNAL_ERROR, str(ex))

        # 判断返回结果，如果不满足要求，则直接返回
        exist_list = tmp_ret and isinstance(tmp_ret, list)
        if exist_list and tmp_ret[0] != 0 and tmp_ret[0] != CommonMethods.OK:
            err_msg = f"{model_name} call {request_type} function failed, function result invalid, message: {tmp_ret}."
            run_log.error(err_msg)
            return AppCommonMethod.get_json_error_by_array(tmp_ret)

        run_log.info(f"{model_name} call {request_type} function end.")

        try:
            if model_name == "System":
                # 获取成功，需要更新缓存
                ret = LibAdapter.get_resource_info(model_name, "all", need_list, False)
                if ret["status"] == CommonMethods.OK:
                    # 获取成功，需要对数据进行缓存
                    key = LibAdapter.get_resource_key("all", need_list)
                    LibAdapter.set_resource(ret["message"], model_name, key, True)
                    run_log.info("Set System resource successfully! ")
        except Exception as e:
            run_log.error(f"Update {model_name} data error! {e}")
        # 进行数据处理
        return CommonMethods.object_to_json(CommonMethods.OK, AppCommonMethod.get_json_info(class_obj))

    @staticmethod
    def lib_timer_interface(model_name, func_key, need_list=False, p_path=None,
                            c_path=None,
                            item1=None, item2=None, item3=None, item4=None,
                            cycles_times=0):
        """
        功能描述：底层接口的定时器调用接口
        参数：modelName 模块名称，用于加载具体的对象
            funcKey 函数关键字
            needList 是否是一个列表，True 或 False
            p_path 父资源的路径，如果发生变动时，需要按照父节点来上报
            cPath 子资源的路径，删除资源时，需要同步删除子节点的信息
            item1 第一级资源
            item2 第二级资源
            item3 第三级资源
            item4 第四级资源
            CyclesTimes 循环次数，用于底层进行数据缓存
        返回值：获取到的信息
        异常描述：NA
        修改记录：2016-11-08 创建
                2016-12-02 调整参数，调用存储接口，存储资源信息；
                调用比较接口判断资源信息是否有变化
                2017-01-23 增加底层数据缓存
        """
        if need_list:
            func_key = "all"
        else:
            # 获取 All 节点缓存，如果为 None，则把 funcKey 设置为 All
            resource = LibAdapter.get_exist_resource(model_name, "all", item1, item2, item3, item4)
            if resource is None:
                func_key = "all"

        ret = LibAdapter.get_resource_info(model_name, func_key, need_list, False,
                                           item1, item2, item3,
                                           item4, cycles_times)
        if ret["status"] != CommonMethods.OK:
            # 资源信息获取失败，直接返回
            return ret

        # 获取成功，需要对数据进行缓存
        key = LibAdapter.get_resource_key(func_key, need_list)
        LibAdapter.set_resource(ret["message"], model_name, key, True, p_path,
                                c_path, item1, item2, item3, item4)

        return ret

    @staticmethod
    def get_resource_info(model_name, func_key, need_list=False, need_cache=True,
                          item1=None, item2=None, item3=None, item4=None,
                          cycles_times=0):
        """
        功能描述：底层接口的定时器调用接口
        参数：modelName 模块名称，用于加载具体的对象
            funcKey 函数关键字
            needList 是否是一个列表，True 或 False
            needCache 是否需要从缓存中获取数据
            item1 第一级资源
            item2 第二级资源
            item3 第三级资源
            item4 第四级资源
            CyclesTimes 循环次数，用于底层进行数据缓存
        返回值：获取到的信息
        异常描述：NA
        修改记录：2016-12-04 创建
                2017-01-14 获取缓存成功后，需要删除变更中的事件信息
                2017-01-23 增加底层数据缓存
        """
        source_key = LibAdapter.get_key_name(model_name, item1, item2, item3, item4)

        c_methods = CommonMethods
        if model_name is None or model_name not in LibAdapter.iBMAClassPath:
            return c_methods.object_to_json(c_methods.INTERNAL_ERROR, "model name is not found.")

        if not LibAdapter.iBMAClassPath.get(model_name) or "keys" not in LibAdapter.iBMAClassPath.get(model_name):
            return c_methods.object_to_json(c_methods.INTERNAL_ERROR, "parameter keys of model is not found.")

        if need_cache:
            # 资源列表中存在，获取缓存数据并返回
            key = LibAdapter.get_resource_key(func_key, need_list)
            e_resource = LibAdapter.get_exist_resource(model_name, key, item1, item2, item3, item4)
            if e_resource is not None:
                # 获取成功需要删除事件列表中的记录，同时删除添加、更新列表中的事件
                LibAdapter.del_event_resource(model_name, RESOURCE_ADDED, item1, item2, item3, item4)
                LibAdapter.del_event_resource(model_name, RESOURCE_UPDATED, item1, item2, item3, item4)
                return c_methods.object_to_json(c_methods.OK, e_resource)

        func_name = f"get_{func_key}_info"
        # 获取底层接口函数
        try:
            tmp_ret = LibAdapter.get_function(LibAdapter.iBMAClassPath[model_name]["class"],
                                              func_name, source_key, not need_cache)
        except Exception as ex:
            run_log.error(f"{source_key} get function failed.")
            return c_methods.object_to_json(c_methods.INTERNAL_ERROR, str(ex))

        if not isinstance(tmp_ret, list):
            return tmp_ret

        class_obj = tmp_ret[0]
        setattr(class_obj, "_local__bdfKeys", LibAdapter.iBMAClassPath.get(model_name).get("keys"))
        # 设置循环次数
        setattr(class_obj, "_local__CyclesTimes", cycles_times)

        # 调用底层接口，获取返回值
        try:
            tmp_ret = LibAdapter.call_function(tmp_ret[1], FuncParams(None, item1, item2, item3, item4))
        except Exception as ex:
            # 需要记录日志信息
            run_log.error(f"{source_key} get information failed.")
            return c_methods.object_to_json(c_methods.INTERNAL_ERROR, str(ex))

        # 判断返回结果，如果不满足要求，则直接返回
        exist_list = tmp_ret and isinstance(tmp_ret, list)
        if exist_list and tmp_ret[0] != 0 and tmp_ret[0] != c_methods.OK:
            if tmp_ret[0] != c_methods.NOT_EXIST:
                # 特殊场景日志屏蔽:任意报文只要token是无效的，导致操作日志被覆盖，影响机密性和完整性、导致无法审计问题
                if source_key or 'Session not found.' not in str(tmp_ret):
                    run_log.error(f"{source_key} get information fail, message:{tmp_ret}.")
            return AppCommonMethod.get_json_error_by_array(tmp_ret)

        if isinstance(need_list, bool) and need_list:
            # 资源列表
            return c_methods.object_to_json(c_methods.OK, getattr(class_obj, "items"))
        # 进行数据转换，将对象的属性转换为JSON
        return c_methods.object_to_json(c_methods.OK, AppCommonMethod.get_json_info(class_obj))

    @staticmethod
    def get_resource_key(func_key, need_list):
        """
        功能描述：获取资源存放的键值
        参数：funcKey 函数关键字
            needList 是否是一个列表，True 或 False
        返回值：资源键值
        异常描述：NA
        修改记录：2016-12-07 创建
        """
        key = func_key
        if need_list:
            # 需要列表类的资源，存放在 items 目录下；其他的资源存放在各自的关键字下
            key = "items"
        return key

    @staticmethod
    def get_ibma_class_path():
        """
        功能描述：返回资源信息清单
        参数：无
        返回值：无
        异常描述：NA
        修改记录：2016-11-30 创建
        """
        return LibAdapter.iBMAClassPath

    @staticmethod
    def get_ibma_resources_value(params):
        """
        功能描述：获取参数列表
        参数：params 获得参数的路径
        返回值：无
        异常描述：NA
        修改记录：2016-11-30 创建
                2017-01-14 修改资源查询方式
                2017-01-18 参数判断时，把键值和类名分开
        """
        # `not` on empty dict should return True
        if params is None or not LibAdapter.iBMAResources:
            return []

        param_list = str(params).split(",")
        # 对应：modelName, item1, item2, item3, item4
        keys = [None, None, None, None, None]
        if len(param_list) < 2:
            return []

        # 模型，目前只有两个参数可配置
        keys[0] = param_list[0]
        # 获取的类型
        type_params = param_list[1]
        key = LibAdapter.get_key_name(keys[0], keys[1], keys[2], keys[3], keys[4])

        # 得到对象所处的模型数据
        model_resources = LibAdapter.iBMAResources.get(key, None)
        if model_resources is None:
            return []

        return model_resources.get(type_params, None)

    @staticmethod
    def get_function(class_path, func_name, key=None, need_cache=False):
        """
        功能描述：获取键值
        参数：classPath 类对象
            funcName 方法名称
            key 关键键值
            needCache 是否需要从缓存中获取数据
        返回值：函数对象
        异常描述：NA
        修改记录：2016-11-14 创建
                2017-01-19 增加缓存
                2017-02-26 增加对对象的初始化，避免出现key不存在的情况
                2017-03-02 将初始化放在互斥锁之外

        # funcName有2类：
        1.startrequest,stoprequest,postrequest,patchrequest
        2.get{}info
        """
        need_init_class = False
        class_object = None
        class_obj = None

        try:
            LibAdapter.mutex.acquire()

            if class_path not in LibAdapter.iBMAClassObjs:
                index = class_path.rfind(".")
                class_name = class_path[index + 1:]
                cls_path = class_path[: index]

                LibAdapter.iBMAClassObjs[class_path] = {
                    "className": class_name, "classPath": cls_path,
                    "classModel": None, "classObj": None
                }

                class_model = importlib.import_module(cls_path)
                if class_model is None or not hasattr(class_model, class_name):
                    LibAdapter.iBMAClassObjs[class_path]["classModel"] = None
                    return CommonMethods.object_to_json(CommonMethods.NOT_EXIST,
                                                        "Class name:" + class_name + " is not found.")
                else:
                    LibAdapter.iBMAClassObjs[class_path]["classModel"] = class_model

                class_obj = getattr(class_model, class_name)
                if class_obj is None:
                    # 找不到实体类，或者方法
                    LibAdapter.iBMAClassObjs[class_path]["classObj"] = None
                    return CommonMethods.object_to_json(CommonMethods.NOT_EXIST,
                                                        "Function name:" + func_name + " is not found.")

                LibAdapter.iBMAClassObjs[class_path]["classObj"] = class_obj

                # 将申请到的对象保存起来
                if key is not None and need_cache:
                    class_object = class_obj()
                    LibAdapter.iBMAClassObjs[class_path][key] = class_object
                else:
                    need_init_class = True
            else:
                if LibAdapter.iBMAClassObjs[class_path]["classModel"] is None:
                    return CommonMethods.object_to_json(CommonMethods.NOT_EXIST,
                                                        "Class name:" + class_path + " is not found.")

                if LibAdapter.iBMAClassObjs[class_path]["classObj"] is None:
                    return CommonMethods.object_to_json(CommonMethods.NOT_EXIST,
                                                        "Function name:" + func_name + " is not found.")

                if key is not None and need_cache:
                    if key in LibAdapter.iBMAClassObjs[class_path]:
                        class_object = LibAdapter.iBMAClassObjs[class_path][key]
                    else:
                        class_obj = LibAdapter.iBMAClassObjs[class_path]["classObj"]
                        class_object = class_obj()
                        LibAdapter.iBMAClassObjs[class_path][key] = class_object
                else:
                    class_obj = LibAdapter.iBMAClassObjs[class_path]["classObj"]
                    need_init_class = True
        finally:
            LibAdapter.mutex.release()

        # 将非缓存的初始化放到互斥锁之外
        if need_init_class and class_obj is not None:
            class_object = class_obj()

        if class_object is None or not hasattr(class_object, func_name):
            # 找不到实体类，或者方法
            return CommonMethods.object_to_json(CommonMethods.NOT_EXIST, f"Function name: {func_name} is not found.")

        func = getattr(class_object, func_name)

        return [class_object, func]

    @staticmethod
    def call_function(func, arg):
        if not func or not callable(func):
            raise ValueError("func is not a function")

        item1, item2, item3, item4 = arg.item1, arg.item2, arg.item3, arg.item4
        if item1 is None:
            return func()
        elif item2 is None:
            return func(item1)
        elif item3 is None:
            return func(item1, item2)
        elif item4 is None:
            return func(item1, item2, item3)
        else:
            return func(item1, item2, item3, item4)

    @staticmethod
    def call_function_with_request_data(func, arg):
        if not func or not callable(func):
            raise ValueError("func is not a function")

        request_data, item1, item2, item3, item4 = arg.request_data, arg.item1, arg.item2, arg.item3, arg.item4
        if item1 is None:
            return func(request_data)
        elif item2 is None:
            return func(request_data, item1)
        elif item3 is None:
            return func(request_data, item1, item2)
        elif item4 is None:
            return func(request_data, item1, item2, item3)
        else:
            return func(request_data, item1, item2, item3, item4)

    @staticmethod
    def delete_class_object_by_items(model_name, item1, item2, item3, item4):
        """
        功能描述：删除资源缓存对象
        参数：modelName 类对象
            item1 第一级资源
            item2 第二级资源
            item3 第三级资源
            item4 第四级资源
        返回值：资源信息
        异常描述：NA
        修改记录：2017-01-25 创建
        """
        if model_name is None or model_name not in LibAdapter.iBMAClassPath:
            return

        source_key = LibAdapter.get_key_name(model_name, item1, item2,
                                             item3, item4)
        class_path = LibAdapter.iBMAClassPath[model_name]["class"]

        try:
            LibAdapter.mutex.acquire()

            if class_path in LibAdapter.iBMAClassObjs \
                    and source_key in LibAdapter.iBMAClassObjs[class_path]:
                run_log.info("%s delete resource %s." % (model_name, source_key))
                LibAdapter.iBMAClassObjs[class_path].pop(source_key)
        finally:
            LibAdapter.mutex.release()

    @staticmethod
    def get_exist_resource(model_name, key, item1, item2, item3, item4):
        """
        功能描述：获取已经存在的资源信息
        参数：modelName 类对象
            key 获取资源的关键字
            item1 第一级资源
            item2 第二级资源
            item3 第三级资源
            item4 第四级资源
        返回值：资源信息
        异常描述：NA
        修改记录：2016-12-04 创建
                2017-01-14 调整资源信息的获取方式
        """
        if model_name not in LibAdapter.ResourceLock:
            return None

        # 获取key值
        model_item_key = LibAdapter.get_key_name(model_name, item1, item2, item3, item4)

        try:
            # 获取资源信息时，需要添加资源锁进行互斥，避免出现不一致的问题
            with LibAdapter.ResourceLock.get(model_name):
                rs = LibAdapter.iBMAResources.get(model_item_key, None)
                if rs is None:
                    return None

                return rs.get(key, None)
        except Exception:
            # 需要记录日志信息
            run_log.error("Get exist resource failed.")
            return None

    @staticmethod
    def set_resource(resource, model_name, key, send_event=True,
                     p_path=None, c_path=None,
                     item1=None, item2=None, item3=None, item4=None):
        """
        功能描述：获取已经存在的资源信息
        参数：resource 资源信息
            modelName 类对象
            key 获取资源的关键字
            sendEvent 是否发送变动事件
            p_path 父资源的路径，如果发生变动时，需要按照父节点来上报
            cPath 子资源的路径，删除资源时，需要同步删除子节点的信息
            item1 第一级资源
            item2 第二级资源
            item3 第三级资源
            item4 第四级资源
        返回值：资源信息
        异常描述：NA
        修改记录：2016-12-05 创建
                2017-01-14 调整资源列表的信息
                2017-02-16 增加异常捕获，防止引起定时器异常退出
        """
        # 互斥锁不存在，说明不支持资源信息缓存
        if model_name not in LibAdapter.ResourceLock:
            return

        # 如果子资源是从All获取的，而本身不是，则需要进行特殊的判断来确认资源是否有变动
        old_is_all = False

        all_resource = None
        # 为了避免出现互斥锁问题，先获取已有的资源信息
        old_resource = LibAdapter.get_exist_resource(model_name, key, item1, item2, item3, item4)
        if key != "all" or key != "items":
            all_resource = LibAdapter.get_exist_resource(model_name, "all", item1, item2, item3, item4)
            if old_resource is None and all_resource is not None:
                # 只有获取到数据时，才认为是从 All 中获取的，
                # 否则认为是从其他的位置获取的
                old_resource = all_resource
                old_is_all = True
        try:
            # 获取资源信息时，需要添加资源锁进行互斥，避免出现不一致的问题
            with LibAdapter.ResourceLock.get(model_name):
                # 获取key值
                model_item_key = LibAdapter.get_key_name(model_name, item1, item2,
                                                         item3, item4)
                if model_item_key not in LibAdapter.iBMAResources:
                    # 没有找到则创建信息
                    LibAdapter.iBMAResources[model_item_key] = {}

                LibAdapter.iBMAResources[model_item_key][key] = resource
        except Exception:
            # 需要记录日志信息
            run_log.error("Set resource failed.")
            return

        try:
            # 将发送，提取到资源锁之外，防止死锁
            if send_event:
                LibAdapter.generate_event(model_name, old_resource, resource,
                                          all_resource, old_is_all, p_path, c_path,
                                          item1, item2, item3, item4)
        except Exception:
            # 需要记录日志信息
            run_log.error("Generate event failed.")

    @staticmethod
    def generate_event(model_name, old_resource, resource,
                       all_resource, old_is_all, p_path=None, c_path=None,
                       item1=None, item2=None, item3=None, item4=None):
        """
        功能描述：判断资源有变化，调用订阅接口，上报变化
        参数：modelName 类对象
            oldResource 原始资源信息
            resource 资源信息
            allResource 所有资源信息
            key 获取资源的关键字
            oldIsAll 原始数据是否为 All 数据
            p_path 父资源的路径，如果发生变动时，需要按照父节点来上报
            cPath 子资源的路径，删除资源时，需要同步删除子节点的信息
            item1 第一级资源
            item2 第二级资源
            item3 第三级资源
            item4 第四级资源
        返回值：资源信息
        异常描述：NA
        修改记录：2016-12-07 创建
                2017-01-15 添加类的资源，需要先获取，等待缓存完成后，
                再上报事件变化
        """
        ret = LibAdapter.check_resources(old_resource, resource, old_is_all)
        if ret[0] != 1:
            return

        change_msg = ret[1]

        # 设置的事件通知的类型，如果存在父路径，说明需要通过父对象上报通知
        e_model_name = model_name
        if p_path is not None:
            e_model_name = p_path

        if isinstance(change_msg, list):
            # 变动为列表，需要遍历上报信息
            a_list = change_msg[1]
            for index, _ in enumerate(a_list):
                c_id = a_list[index]
                run_log.info(f"{e_model_name} has changed, add source:{c_id}.")

            d_list = change_msg[0]
            for index, _ in enumerate(d_list):
                c_id = d_list[index]
                # 删除子节点的数据
                run_log.info(f"{e_model_name} has changed, remove source:{c_id}.")
                LibAdapter.delete_resource(c_path, c_id, item1, item2, item3, item4)
                # 删除自身节点数据
                LibAdapter.delete_resource(model_name, c_id, item1, item2, item3, item4)
                # 上报变更事件

            return

        # 资源发生变化
        run_log.info(f"{model_name} resource has changed, update source:{change_msg}.")

        if old_is_all:
            # 需要修改父节点属性信息，但是不检查是否需要更新
            LibAdapter.set_resource(change_msg, e_model_name, "all", False, None, None,
                                    item1, item2, item3, item4)
        else:
            # 只有部分属性发生了变化，只需要更新All中的部分属性
            if all_resource is not None:
                # 替换掉 all 中的属性值
                change_msg = LibAdapter.replace_change_attributes(all_resource, change_msg)
                if change_msg:
                    # 设置 All 属性
                    LibAdapter.set_resource(change_msg, e_model_name, "all", False,
                                            None, None, item1, item2, item3,
                                            item4)

    @staticmethod
    def delete_resource(c_path, c_id, item1=None, item2=None, item3=None,
                        item4=None):
        """
        功能描述：删除指定的资源信息
        参数：cPath 子资源名称
            cID 变化的资源id
            item1 第一级资源
            item2 第二级资源
            item3 第三级资源
            item4 第四级资源
        返回值：资源信息
        异常描述：NA
        修改记录：2016-12-07 创建
                2016-12-16 调整删除代码，增加对子节点路径的控制
                2017-01-14 调整键值的方式
        """
        if c_path is None:
            return

        # 使用 cID 替换掉第一个为 None 的参数
        if item1 is None:
            item1 = c_id
        elif item2 is None:
            item2 = c_id
        elif item3 is None:
            item3 = c_id
        elif item4 is None:
            item4 = c_id
        else:
            return

        c_list = c_path.split(";")
        try:
            for index, _ in enumerate(c_list):
                c_model_name = c_list[index]

                # 获取key值
                key = LibAdapter.get_key_name(c_model_name, item1, item2, item3, item4)

                # 删除资源
                LibAdapter.iBMAResources.pop(key, "")

                # 删除缓存中的对象
                LibAdapter.delete_class_object_by_items(c_model_name, item1, item2, item3, item4)
        except Exception:
            run_log.error("delete resource failed.")

    @staticmethod
    def check_resources(old_resource, resource, old_is_all=False):
        """
        功能描述：获取已经存在的资源信息
        参数：oldResource 原有资源信息
            resource 最新资源信息
            oldIsAll 旧的资源信息是 All，而新资源是动态的，
            此时比较肯定会不一致，只判断新资源中值不为 None 的项
        返回值：0：没有变化
            1：有变化，如果资源是列表的话，则需要返回添加、删除的资源
        异常描述：NA
        修改记录：2016-12-06 创建
        """
        ret = [0, ""]

        if old_resource is None and not old_is_all:
            # 如果原有的数据为空，则直接保存所有的信息，
            # 认为尚未初始化完成，不上报变动事件
            ret[1] = resource
            return ret

        if resource is None:
            # 获取到的数据为 None，说明下层处理可能出现问题，此处不做处理，认为没有变化
            return ret

        if isinstance(resource, list):
            # 对列表元素进行比较，先转成集合
            old_resource_set = set(old_resource)
            resource_set = set(resource)
            # 转成集合，然后判断是否相等
            if old_resource_set == resource_set:
                return ret

            c_list = [[], []]
            # 得到删除的列表
            c_list[0] = list(old_resource_set.difference(resource_set))
            # 得到增加的列表
            c_list[1] = list(resource_set.difference(old_resource_set))

            ret[0] = 1
            ret[1] = c_list
        elif isinstance(resource, dict):
            # 对字典的属性进行比较
            if old_is_all:
                ret = LibAdapter.check_dict(old_resource, resource, old_is_all)
            elif operator.eq(old_resource, resource) != 0:
                # 之前是py2的cwp函数
                # 有变化，设置
                ret = LibAdapter.check_dict(old_resource, resource, old_is_all)
        else:
            # 返回的类型有问题
            ret[0] = -1
            ret[1] = "Error resource type."

        return ret

    @staticmethod
    def check_dict(old_resource, resource, old_is_all=False):
        """
        功能描述：判断两个字典是否相等
        参数：oldResource 原有资源信息
            resource 最新资源信息
            oldIsAll 旧的资源信息是 All，而新资源是动态的，此时比较肯定会不一致，只判断新资源中值不为 None 的项
        返回值：0：没有变化
            1：有变化，如果资源是列表的话，则需要返回添加、删除的资源
        异常描述：NA
        修改记录：2016-12-06 创建
        """
        ret = [0, {}]
        # 保存变化的属性
        c_resource = {}

        # 遍历新资源的每个属性，判断是否一致
        for key in resource.keys():
            new_value = resource[key]

            if key not in old_resource:
                # 有变化，如果老节点中没有找到对应信息，则认为发生了变化
                ret[0] = 1
                old_resource[key] = new_value
                c_resource[key] = new_value
                continue

            old_value = old_resource[key]
            if new_value is None:
                if old_is_all:
                    # 由于新获取到的资源，没有原始的资源作为对比，所以认为新资源中为 None 的属性没有发生变化
                    continue
                else:
                    if old_value is not None:
                        # 发生变化
                        ret[0] = 1
                        old_resource[key] = new_value
                        c_resource[key] = new_value
                        continue
            else:
                if old_value is None:
                    # 发生变化
                    ret[0] = 1
                    old_resource[key] = new_value
                    c_resource[key] = new_value
                    continue

            # 两者都不为空，则需要通过类型的方式判断具体的是否有变化
            if isinstance(old_value, list) and isinstance(new_value, list):
                if not new_value and old_is_all:
                    # 认为没有发生变化
                    continue

                # 转成集合，然后判断是否相等
                if not old_value:
                    if new_value:
                        ret[0] = 1
                        old_resource[key] = new_value
                        c_resource[key] = new_value
                        continue
                else:
                    tmp_ret = LibAdapter.check_list(old_value, new_value)
                    if tmp_ret[0] != 0:
                        ret[0] = tmp_ret[0]
                        old_resource[key] = new_value
                        c_resource[key] = new_value
            elif isinstance(old_value, dict) and isinstance(new_value, dict):
                if not new_value and old_is_all:
                    # 认为没有发生变化
                    continue

                tmp_ret = LibAdapter.check_dict(old_value, new_value, old_is_all)
                if tmp_ret[0] != 0:
                    ret[0] = tmp_ret[0]
                    old_resource[key] = new_value
                    c_resource[key] = new_value
            else:
                # 其他情况转换成字符串进行比较，如果不一样，则认为发生了变化
                if str(old_value) != str(new_value):
                    ret[0] = 1
                    old_resource[key] = new_value
                    c_resource[key] = new_value

        if old_is_all:
            ret[1] = old_resource
        else:
            ret[1] = c_resource

        return ret

    @staticmethod
    def check_list(old_value, new_value):
        """
        功能描述：判断两个字典是否相等
        参数：oValue 原有资源信息
            nValue 最新资源信息
        返回值：0：没有变化
            1：有变化，如果资源是列表的话，则需要返回添加、删除的资源
        异常描述：NA
        修改记录：2016-12-06 创建
        """
        ret = [0, {}]
        if len(old_value) != len(new_value):
            ret[0] = 1
            ret[1] = new_value
            return ret

        old_value.sort()
        new_value.sort()

        for index, _ in enumerate(old_value):
            oi = old_value[index]
            ni = new_value[index]
            if isinstance(oi, list) and isinstance(ni, list):
                tmp_ret = LibAdapter.check_list(oi, ni)
                if tmp_ret[0] != 0:
                    ret[0] = tmp_ret[0]
                    old_value = new_value
                    break
            elif isinstance(oi, dict) and isinstance(ni, dict):
                tmp_ret = LibAdapter.check_dict(oi, ni)
                if tmp_ret[0] != 0:
                    ret[0] = tmp_ret[0]
                    old_value = new_value
                    break
            else:
                # 其他情况转换成字符串进行比较，如果不一样，则认为发生了变化
                if str(oi) != str(ni):
                    ret[0] = 1
                    old_value = new_value
                    break

        ret[1] = old_value

        return ret

    @staticmethod
    def del_event_resource(model_name, event_type, item1=None,
                           item2=None, item3=None, item4=None):
        """
        功能描述：检查是否在列表中，如果在就删除掉
        参数：modelName 类对象
            eventType 资源变化类型
            item1 第一级资源
            item2 第二级资源
            item3 第三级资源
            item4 第四级资源
        返回值：无
        异常描述：NA
        修改记录：2017-01-14 创建
        """
        event_key = LibAdapter.get_key_name(model_name, item1, item2, item3, item4)
        try:
            with LibAdapter.eventResourcesLock:
                if event_key in LibAdapter.eventResourcesDict[event_type]:
                    LibAdapter.eventResourcesDict[event_type].pop(event_key)
        except Exception:
            run_log.error(f"delete event resource failed.")

    @staticmethod
    def get_key_name(model_name, item1=None, item2=None, item3=None, item4=None):
        """
        功能描述：获取关键key值
        参数：modelName 类对象
            item1 第一级资源
            item2 第二级资源
            item3 第三级资源
            item4 第四级资源
        返回值：无
        异常描述：NA
        修改记录：2017-01-14 创建
                2017-03-02 调整创建的key值
        """
        return f"{model_name}_{item1}_{item2}_{item3}_{item4}"

    @staticmethod
    def replace_change_attributes(all_resource, msg):
        """
        功能描述：替换 All 中的属性值
        参数：allResource 原有资源信息
            cMsg 修改的资源信息
        返回值：替换后的资源
        异常描述：NA
        修改记录：2016-12-09 创建
        """
        if all_resource is None or msg is None:
            return {}

        if not isinstance(all_resource, dict) or not isinstance(msg, dict):
            # 需要两者均为字典，否则认为没有变化
            return {}

        for key in msg.keys():
            all_resource[key] = msg[key]

        return all_resource

    @staticmethod
    def set_ibma_timers(timers):
        """
        功能描述：设置定时器
        参数：timers 定时器信息
        返回值：无
        异常描述：NA
        修改记录：2017-01-20 创建
        """
        LibAdapter.ibma_timers = timers

    @staticmethod
    def get_ibma_timers_cfg():
        """
        功能描述：获取定时器配置文件
        参数：无
        返回值：无
        异常描述：NA
        修改记录：2017-01-20 创建
        """
        return LibAdapter.ibma_timers_cfg

    @staticmethod
    def set_ibma_timers_cfg(timers_cfg):
        """
        功能描述：设置定时器配置文件
        参数：timersCfg 定时器配置文件信息
        返回值：无
        异常描述：NA
        修改记录：2017-01-20 创建
        """
        LibAdapter.ibma_timers_cfg = timers_cfg
