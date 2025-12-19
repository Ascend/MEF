# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.


class DeviceManagerError(Exception):
    pass


class UnsupportedDeviceAttrTypeError(DeviceManagerError):
    """不支持的设备属性类型"""


class TooManyModuleFilesError(DeviceManagerError):
    """模组配置文件达到上限"""


class ValueLengthExceedsLimit(DeviceManagerError):
    """val值转换为byte长度超过"""


class DeviceNotExistError(DeviceManagerError):
    """
    This is when you pass a device name to DeviceManager.get_device
    that doesn't exist in json specification
    """


class ModuleNotExistError(DeviceManagerError):
    """
    This is when you call DEVM.get_device_list_by_module or DEVM.get_module_attributes with a non-exist module name.
    """


class AttributeNotExistError(DeviceManagerError):
    """
    This is when you pass an attribute name to Device.get_attribute or Device.set_attribute
    that doesn't exist in json specification
    """


class MethodNotAllowedError(DeviceManagerError):
    """
    This is when you call set_attribute/get_attribute on Device
    but the attribute's accessMode doesn't support the corresponding operation
    """


class InconsistentDateTypeError(DeviceManagerError):
    """
    This is when you're trying to set an attribute with a wrong data type
    """


class DriverError(DeviceManagerError):
    """
    This is non-zero result is received by calling driver function
    """
