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
import struct
from abc import ABCMeta
from abc import abstractmethod
from typing import Dict
from typing import Tuple
from typing import Type

from common.log.logger import run_log
from devm import config
from devm.constants import DeviceAttrAccessModeEnum
from devm.constants import DeviceAttrTypeEnum
from devm.constants import DevmValType
from devm.exception import AttributeNotExistError
from devm.exception import DeviceManagerError
from devm.exception import UnsupportedDeviceAttrTypeError
from devm.exception import ValueLengthExceedsLimit


class StrUtil:
    @staticmethod
    def str_to_bytes(data) -> bytes:
        """将data数据转换为bytes类型"""
        if isinstance(data, bytes):
            return data
        if isinstance(data, str):
            return data.encode(encoding="utf-8")
        run_log.error("Value must be represented as bytes or unicode string")
        raise DeviceManagerError("Value must be represented as bytes or unicode string")

    @staticmethod
    def load_from_bytes(byte_data: bytes) -> str:
        """
        将字节数组转换为字符串，有效数据为字节数组开始到首个'\x00'字节位置，若没有'\x00'字节则返回空字符串
        :param byte_data: 以'\x00'结尾的字节数组
        :return: 字符串
        """
        end_char_idx = byte_data.find(b"\x00")
        return str(byte_data[:end_char_idx], encoding="utf-8") if end_char_idx > 0 else ""


class DeviceAttr(metaclass=ABCMeta):
    def __init__(self, name, attr_info):
        self.name: str = name
        self.id: int = attr_info["id"]
        self.type = DeviceAttrTypeEnum(attr_info["type"])
        self.access_mode = DeviceAttrAccessModeEnum(attr_info["accessMode"])

    @staticmethod
    def load_tag(tag_buff: bytes) -> int:
        return struct.unpack("=i", tag_buff)[0]

    @staticmethod
    def load_len(len_buff: bytes) -> int:
        return struct.unpack("=i", len_buff)[0]

    @abstractmethod
    def get_val_len(self, val=None) -> int:
        """
        获取Value的tlv格式长度

        :return:
        """
        pass

    @abstractmethod
    def dump_val(self, val) -> bytes:
        """
        将val序列化为二进制

        :param val:
        :return:
        """
        pass

    @abstractmethod
    def load_val(self, val_tlv: bytes) -> DevmValType:
        """
        将二进制的val反序列化

        :param val_tlv:
        :return:
        """
        pass

    def dump_tag(self) -> bytes:
        return struct.pack("=i", self.id)

    def dump_len(self, val=None) -> bytes:
        return struct.pack("=i", self.get_val_len(val))

    def dump_tlv(self, val=None) -> bytes:
        return b"".join([self.dump_tag(), self.dump_len(val), self.dump_val(val)])

    def load_tlv(self, buff: bytes) -> Tuple[int, int, DevmValType]:
        if len(buff) > config.MAX_BUFFER_SIZE_BYTE:
            run_log.error("tlv buffer bytes length exceeds limit, attr name is %s", self.name)
            raise ValueLengthExceedsLimit(f"tlv buffer bytes length exceeds limit, attr name is {self.name}")

        type_byte_size = ctypes.sizeof(ctypes.c_int)
        tag_buff = buff[:type_byte_size]
        pos = type_byte_size * 2
        len_buff = buff[type_byte_size:pos]
        val_buff = buff[pos:]
        return self.load_tag(tag_buff), self.load_len(len_buff), self.load_val(val_buff)

    def calc_required_buff_size(self, val=None) -> int:
        return ctypes.sizeof(ctypes.c_int) * 2 + self.get_val_len(val)

    def to_json(self) -> Dict:
        return {"accessMode": self.access_mode.value, "valueType": self.type.value}


class DeviceAttrInt(DeviceAttr):
    def get_val_len(self, val=None) -> int:
        return ctypes.sizeof(ctypes.c_int)

    def dump_val(self, val) -> bytes:
        if not val:
            return b"\x00" * self.get_val_len()
        return struct.pack("=i", val)

    def load_val(self, val_tlv: bytes) -> int:
        return struct.unpack("=i", val_tlv)[0]


class DeviceAttrFloat(DeviceAttr):
    def get_val_len(self, val=None) -> int:
        return ctypes.sizeof(ctypes.c_float)

    def dump_val(self, val) -> bytes:
        if not val:
            return b"\x00" * self.get_val_len()
        return struct.pack("=f", val)

    def load_val(self, val_tlv: bytes) -> float:
        return struct.unpack("=f", val_tlv)[0]


class DeviceAttrStr(DeviceAttr):
    STRING_SIZE = config.MAX_STRING_BUFFER_SIZE_BYTE

    def get_val_len(self, val=None) -> int:
        return self.STRING_SIZE

    def dump_val(self, val) -> bytes:
        if not val:
            return b"\x00" * self.get_val_len()
        val_bytes = StrUtil.str_to_bytes(val)
        if len(val_bytes) > self.STRING_SIZE:
            run_log.error("string value length exceeds limit, attribute name is %s", self.name)
            raise ValueLengthExceedsLimit(f"string value length exceeds limit, attribute name is {self.name}")

        return struct.pack(f"={self.get_val_len()}s", val_bytes)

    def load_val(self, val_tlv: bytes) -> str:
        return StrUtil.load_from_bytes(val_tlv)


class DeviceAttrLongLong(DeviceAttr):
    def get_val_len(self, val=None) -> int:
        return ctypes.sizeof(ctypes.c_ulonglong)

    def dump_val(self, val) -> bytes:
        if not val:
            return b"\x00" * self.get_val_len()
        return struct.pack("=Q", val)

    def load_val(self, val_tlv: bytes) -> int:
        return struct.unpack("=Q", val_tlv)[0]


class DeviceAttrJson(DeviceAttr):
    def __init__(self, name, attr_info):
        super().__init__(name, attr_info)

        self.sub_attrs: Dict[str, DeviceAttr] = {}
        for sub_attr_name, sub_attr in attr_info.get("subAttributes", {}).items():
            attr_type = DeviceAttrTypeEnum(sub_attr["type"])
            if attr_type == DeviceAttrTypeEnum.JSON:
                continue
            self.sub_attrs[sub_attr_name] = attr_factory(attr_type, sub_attr_name, sub_attr)

    def get_val_len(self, val=None) -> int:
        sub_attrs, val = self._get_sub_attrs_by_val(val)
        return sum(sub_attr.calc_required_buff_size(val.get(name)) for name, sub_attr in sub_attrs.items())

    def dump_val(self, val) -> bytes:
        bytes_list = []
        val_buff_size = 0
        max_val_byte = config.MAX_BUFFER_SIZE_BYTE - config.HEADER_BUFF_SIZE_BYTE
        sub_attrs, val = self._get_sub_attrs_by_val(val)
        for sub_attr_name, sub_attr in sub_attrs.items():
            tlv_bytes = sub_attr.dump_tlv(val.get(sub_attr_name))
            val_buff_size += len(tlv_bytes)
            if val_buff_size > max_val_byte:
                run_log.error("json value length exceeds limit, attribute name is %s", self.name)
                raise ValueLengthExceedsLimit(f"json value length exceeds limit, attribute name is {self.name}")

            bytes_list.append(tlv_bytes)

        return b"".join(bytes_list)

    def load_val(self, val_tlv: bytes) -> Dict:
        len_val = len(val_tlv)
        if len_val > config.MAX_BUFFER_SIZE_BYTE:
            run_log.error("tlv buffer size is too large, must be less than or equal to %s", config.MAX_BUFFER_SIZE_BYTE)
            raise UnsupportedDeviceAttrTypeError("tlv buffer size is too large")

        tag_len = ctypes.sizeof(ctypes.c_int)
        val = {}
        pos = 0
        while pos < len_val:
            tag_val = struct.unpack("=i", val_tlv[pos:pos + tag_len])[0]
            sub_attr = self._get_sub_attr_by_attr_id(tag_val)
            sub_attr_tlv_len = sub_attr.calc_required_buff_size()
            res = sub_attr.load_tlv(val_tlv[pos:pos + sub_attr_tlv_len])
            pos += sub_attr_tlv_len
            val.update({sub_attr.name: res[2]})

        return val

    def to_json(self) -> Dict:
        json_data = super().to_json()
        sub_attr_json = {sub_attr_name: sub_attr.to_json() for sub_attr_name, sub_attr in self.sub_attrs.items()}
        json_data.update({"subAttributes": sub_attr_json})
        return json_data

    def _get_sub_attr_by_attr_id(self, attr_id) -> DeviceAttr:
        for sub_attr in self.sub_attrs.values():
            if sub_attr.id == attr_id:
                return sub_attr

        run_log.error("invalid attr id: %s", attr_id)
        raise AttributeNotExistError(f"invalid attr id: {attr_id}")

    def _get_sub_attrs_by_val(self, val) -> Tuple[Dict[str, DeviceAttr], Dict[str, DevmValType]]:
        if val and not isinstance(val, dict):
            run_log.error("invalid parameters: value must be empty or dict format")
            raise UnsupportedDeviceAttrTypeError("invalid parameters")

        if not val:
            return self.sub_attrs, {}

        sub_attrs = {}
        for sub_attr_name in val:
            if sub_attr_name not in self.sub_attrs:
                run_log.warning("invalid sub attr name: %s, ignored", sub_attr_name)
                continue

            sub_attrs.update({sub_attr_name: self.sub_attrs[sub_attr_name]})
        return sub_attrs, val


class DeviceAttrBool(DeviceAttr):

    def get_val_len(self, val=None):
        return ctypes.sizeof(ctypes.c_bool)

    def dump_val(self, val) -> bytes:
        if not val:
            return b"\x00" * self.get_val_len()
        return struct.pack("=?", val)

    def load_val(self, val_tlv: bytes):
        return struct.unpack("=?", val_tlv)[0]


IMPLS: Dict[DeviceAttrTypeEnum, Type[DeviceAttr]] = {
    DeviceAttrTypeEnum.INT: DeviceAttrInt,
    DeviceAttrTypeEnum.FLOAT: DeviceAttrFloat,
    DeviceAttrTypeEnum.LONG_LONG: DeviceAttrLongLong,
    DeviceAttrTypeEnum.STRING: DeviceAttrStr,
    DeviceAttrTypeEnum.JSON: DeviceAttrJson,
    DeviceAttrTypeEnum.BOOL: DeviceAttrBool
}


def attr_factory(attr_type: DeviceAttrTypeEnum, attr_name, attr_info: Dict) -> DeviceAttr:
    cls = IMPLS.get(attr_type)
    if not cls:
        run_log.error("unsupported device attr type: %s", attr_type)
        raise UnsupportedDeviceAttrTypeError()

    return cls(attr_name, attr_info)
