# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

import threading

from common.log.logger import run_log
from common.utils.common_check import CommonCheck
from common.common_methods import CommonMethods
from devm.device_mgr import DEVM
from devm.exception import DeviceManagerError


class Module:

    def __init__(self):
        """
        功能描述：初始化函数
        参数：
        返回值：无
        异常描述：NA
        """
        # 初始化模组列表
        self.items = []
        self.devices = []
        self.ModuleInfo = {}
        self.get_module_list()

    def get_all_info(self, module_id=None) -> list:
        """
        功能描述：获取模组资源列表以及指定模组信息
        参数：module_id 模组名 为None是获取模组列表
        返回值：list
        异常描述：NA
        """
        self.get_module_list()
        if module_id is None:
            return [0, ]

        if module_id not in self.items:
            run_log.error("module %s not exist", module_id)
            return [CommonMethods.NOT_EXIST, '{} does not exist'.format(module_id)]

        self.get_module_info(module_id)
        run_log.info("get module %s info success", module_id)
        return [CommonMethods.OK, "get module info success"]

    def get_module_list(self):
        try:
            # 目前获取的模组列表仅包含拓展模组
            self.items = DEVM.get_module_list_by_category('addition')
        except DeviceManagerError:
            run_log.error("failed to get module list.")

    def get_module_info(self, module_id):
        try:
            self.devices = DEVM.get_device_list_by_module(module_id)
        except DeviceManagerError:
            run_log.error("failed to get device list, module name is %s", module_id)
            return

        try:
            self.ModuleInfo = DEVM.get_module_info(module_id)
        except DeviceManagerError:
            run_log.error("failed to get module info, module name is %s", module_id)


class Device:
    DEVICE_LOCK = threading.Lock()

    def __init__(self):
        """
        功能描述：初始化函数
        参数：
        返回值：无
        异常描述：NA
        """
        self.Attributes = {}

    @staticmethod
    def set_device_info(device_id, attributes: dict):
        for attribute_name, value in attributes.items():
            try:
                DEVM.get_device(device_id).set_attribute(attribute_name, value)
            except DeviceManagerError:
                run_log.error("failed to set device %s attribute %s" % device_id, attribute_name)
                return

    @staticmethod
    def _check_if_device_belongs_to_module(device_id, module_id):
        return device_id in DEVM.get_device_list_by_module(module_id)

    def get_all_info(self, module_id=None, device_id=None) -> list:
        """
        功能描述：获取模组资源列表以及指定模组信息
        参数：module_id 模组名 device_id 设备名
        返回值：list
        异常描述：NA
        """
        module_obj = Module()
        if module_id not in module_obj.items:
            run_log.error("module %s not exist", module_id)
            return [CommonMethods.NOT_EXIST, '{} does not exist'.format(module_id)]

        if not self._check_if_device_belongs_to_module(device_id, module_id):
            run_log.error("module %s has no device %s", module_id, device_id)
            return [CommonMethods.NOT_EXIST, f"module '{module_id}' has no device '{device_id}'"]

        self.get_device_info(device_id)
        run_log.info("get device %s attributes success", device_id)
        return [CommonMethods.OK, "get device info success"]

    def patch_request(self, request_data, module_id=None, device_id=None):
        """
        功能描述：patch请求的处理参数
        参数：request_data  请求参数
             module_id为模组名称，device_id为设备名称
        返回值：是否成功的列表信息
        异常描述：
        """
        module_obj = Module()
        if module_id not in module_obj.items:
            run_log.error("module %s not exist", module_id)
            return [CommonMethods.NOT_EXIST, '{} does not exist'.format(module_id)]

        if not self._check_if_device_belongs_to_module(device_id, module_id):
            run_log.error("module %s has no device %s", module_id, device_id)
            return [CommonMethods.NOT_EXIST, f"module '{module_id}' has no device '{device_id}'"]

        if Device.DEVICE_LOCK.locked():
            run_log.warning("Device modify is busy")
            return [CommonMethods.ERROR, "Device modify is busy"]

        with Device.DEVICE_LOCK:
            try:
                operator_check = CommonCheck.check_operator(request_data.get("_User"), request_data.get("_Xip"))
                if not operator_check:
                    run_log.error("The operator is illegal, %s", operator_check.error)
                    return [CommonMethods.ERROR, "The operator is illegal."]

                self.set_device_info(device_id, request_data.get("Attributes", {}))
                run_log.info("set device %s info success, module name : %s", device_id, module_id)
                self.get_device_info(device_id)
                run_log.info("get device %s info success", device_id)
                return [CommonMethods.OK, "Device modify success"]
            except Exception as err:
                run_log.error("Device modify failed, reason is %s", (str(err)))
                return [CommonMethods.ERROR, "Device modify failed, reason is {}".format(str(err))]

    def get_device_info(self, device_id):
        try:
            self.Attributes = DEVM.get_device(device_id).get_all_attributes()
        except DeviceManagerError:
            run_log.error("failed to get device %s attributes." % device_id)
