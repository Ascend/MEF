# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import ctypes
from typing import Set

from common.log.logger import run_log
from devm import config
from devm.device_attrs import StrUtil
from devm.exception import DriverError


def reset_char_buffer(char_p: ctypes.Array[ctypes.c_char]) -> None:
    """内存空间写0"""
    char_p.value = b"\x00" * len(char_p)


class DeviceDriver:
    """提供对驱动so的封装"""

    def __init__(self, so_file_name):
        self.so_file_name = so_file_name
        self.dll_inst: ctypes.CDLL = self._load_so(self.so_file_name)
        self.load()

    @staticmethod
    def _load_so(so_file_name):
        try:
            dll_inst = ctypes.CDLL(so_file_name, mode=ctypes.RTLD_GLOBAL)
        except Exception as err:
            run_log.error("load so %s failed: %s", so_file_name, err)
            raise Exception(f"load {so_file_name} so failed: {err}") from err
        run_log.info("loaded so %s successfully", so_file_name)
        return dll_inst

    def get_func(self, func_name, arg_types, res_type):
        func = getattr(self.dll_inst, func_name)
        func.argtypes = arg_types
        func.restype = res_type
        return func

    def load(self):
        """用于加载驱动，分配模组所需资源，实现模组初始化。"""
        dev_load = self.get_func(func_name="dev_load", arg_types=[], res_type=ctypes.c_int)
        ret = dev_load()
        if ret != 0:
            run_log.warning("call dev_load in %s failed, ret=%s", self.so_file_name, ret)

    def unload(self):
        """用于卸载驱动，释放模组所占资源。"""
        dev_unload = self.get_func(func_name="dev_unload", arg_types=[], res_type=ctypes.c_int)
        ret = dev_unload()
        if ret != 0:
            run_log.warning("call dev_unload in %s failed, ret=%s", self.so_file_name, ret)

    def open(self, device_name) -> int:
        """
        用于打开设备，调用驱动的dev_open函数，根据传入的设备名分配设备文件描述符
        :param device_name: 设备名
        :return: 设备文件描述符
        """
        dev_open = self.get_func(func_name="dev_open", arg_types=[ctypes.c_char_p, ctypes.POINTER(ctypes.c_int)],
                                 res_type=ctypes.c_int)
        device_name_p = ctypes.create_string_buffer(StrUtil.str_to_bytes(device_name))
        device_fd = ctypes.c_int()

        ret = dev_open(device_name_p, ctypes.byref(device_fd))
        if ret != 0:
            run_log.error("call dev_open failed, ret=%s", ret)
            raise DriverError(f"call dev_open failed, ret={ret}")
        return device_fd.value

    def close(self, fd: int):
        """
        用于关闭设备，调用驱动的dev_close函数，根据fd删除对应的设备信息
        :param fd: 设备文件描述符
        :return:
        """
        dev_close = self.get_func(func_name="dev_close", arg_types=[ctypes.c_int], res_type=ctypes.c_int)
        ret = dev_close(ctypes.c_int(fd))
        if ret != 0:
            run_log.error("call dev_close failed, ret=%s", ret)
            raise DriverError(f"call dev_close failed, ret={ret}")
        return ret

    def get_attr(self, fd, tlv: bytes) -> bytes:
        """
        获取属性接口，调用驱动的get_attribute函数，根据传入fd值确定唯一设备，通过TLV编码的buff解析获取属性ID和对应属性值
        :param fd: 设备文件描述符
        :param tlv: 通过TLV编码的buff字节数组
        :return:
        """
        get_attribute = self.get_func(func_name="get_attribute",
                                      arg_types=[ctypes.c_int, ctypes.c_uint, ctypes.c_char_p], res_type=ctypes.c_int)
        tlv_buff_p = ctypes.create_string_buffer(tlv)
        tlv_buff_len = len(tlv)
        ret = get_attribute(ctypes.c_int(fd), ctypes.c_uint(tlv_buff_len), tlv_buff_p)
        if ret != 0:
            reset_char_buffer(tlv_buff_p)
            run_log.error("call get_attribute func failed: ret=%s", ret)
            raise DriverError(f"call get_attribute func failed: ret={ret}")

        result = tlv_buff_p.raw[:-1]
        reset_char_buffer(tlv_buff_p)
        return result

    def set_attr(self, fd, tlv: bytes):
        """
        设置属性接口，调用驱动的set_attribute函数，根据传入fd值确定唯一设备，通过TLV编码的buff解析获取属性id并设置对应属性值
        :param fd: 设备文件描述符
        :param tlv: 通过TLV编码的buff字节数组
        :return:
        """
        set_attribute = self.get_func(func_name="set_attribute",
                                      arg_types=[ctypes.c_int, ctypes.c_uint, ctypes.c_char_p], res_type=ctypes.c_int)
        tlv_buff_p = ctypes.create_string_buffer(tlv)
        tlv_buff_len = len(tlv)
        ret = set_attribute(ctypes.c_int(fd), ctypes.c_uint(tlv_buff_len), tlv_buff_p)
        if ret != 0:
            reset_char_buffer(tlv_buff_p)
            run_log.error("call set_attribute func failed: ret=%s", ret)
            raise DriverError(f"call set_attribute func failed: ret={ret}")

        reset_char_buffer(tlv_buff_p)
        return ret

    def get_dynamic_devices(self, module_id) -> Set[str]:
        """
        动态获取指定类型模组的设备列表
        :param module_id: 获取的设备类型，与模组配置文件中的ID对应
        :return: 设备名称列表
        """
        dev_get_device_list = self.get_func(func_name="dev_get_device_list",
                                            arg_types=[ctypes.c_int, ctypes.c_uint, ctypes.c_char_p, ctypes.c_uint],
                                            res_type=ctypes.c_int)
        buff_size = config.MAX_DYNAMIC_MODULE_DEVICE_BUFFER_SIZE_BYTE
        buff = ctypes.create_string_buffer(buff_size)
        dev_name_len = config.MAX_DYNAMIC_MODULE_DEVICE_NAME_LEN
        ret = dev_get_device_list(ctypes.c_int(module_id), ctypes.c_uint(buff_size), buff, ctypes.c_uint(dev_name_len))
        if ret != 0:
            raise DriverError(f"call dev_get_device_list func failed: ret={ret}")

        dev_names = set()
        pos = 0
        while pos < buff_size:
            name = StrUtil.load_from_bytes(buff.raw[pos: pos + dev_name_len])
            pos += dev_name_len
            if not name:
                continue
            dev_names.add(name)
        return dev_names
