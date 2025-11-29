# coding: utf-8
#  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import struct
from collections import namedtuple
from enum import Enum

import pytest

from devm.config import MAX_STRING_BUFFER_SIZE_BYTE
from devm.constants import DeviceAttrTypeEnum
from devm.device_attrs import DeviceAttr
from devm.device_attrs import DeviceAttrInt
from devm.device_attrs import DeviceAttrJson
from devm.device_attrs import attr_factory
from devm.exception import UnsupportedDeviceAttrTypeError


class InvalidDeviceAttrTypeEnum(Enum):
    UINT = "unsigned int"
    XML = "xml"


class TestDeviceAttr:
    LoadTagCase = namedtuple("LoadTagCase", ["expect", "tag_buff"])
    GetValueLenCase = namedtuple("GetValueLenCase", ["expect", "attr_type", "attr_name", "attr_info", "value"])
    AttrFactoryCase = namedtuple("AttrFactoryCase", ["expect", "attr_type", "attr_name", "attr_info"])
    use_cases = {
        "test_load_tag_normal": {
            "tag_buff_len_4_should_ok": (1, b"\x01\x00\x00\x00"),
        },
        "test_load_tag_error": {
            "tag_buff_len_0_should_error": (struct.error, b""),
            "tag_buff_len_3_should_error": (struct.error, b"\x01\x00\x00"),
            "tag_buff_len_5_should_error": (struct.error, b"\x01\x00\x00\x00\x00"),
        },
        "test_get_val_len": {
            "int": (4, DeviceAttrTypeEnum.INT, "attr_int", {"id": 1, "type": "int", "accessMode": "ReadWrite"}, None),
            "float": (4, DeviceAttrTypeEnum.FLOAT, "attr_float",
                      {"id": 2, "type": "float", "accessMode": "ReadWrite"}, None),
            "longlong": (8, DeviceAttrTypeEnum.LONG_LONG, "attr_longlong",
                         {"id": 3, "type": "long long", "accessMode": "ReadWrite"}, None),
            "string": (MAX_STRING_BUFFER_SIZE_BYTE, DeviceAttrTypeEnum.STRING, "attr_string",
                       {"id": 4, "type": "string", "accessMode": "ReadWrite"}, None),
            "bool": (1, DeviceAttrTypeEnum.BOOL, "attr_bool",
                     {"id": 5, "type": "bool", "accessMode": "ReadWrite"}, None),
            "json": (9, DeviceAttrTypeEnum.JSON, "attr_json",
                     {
                         "id": 6, "type": "json", "accessMode": "ReadWrite",
                         "subAttributes": {"switch_on": {"id": 10, "type": "bool", "accessMode": "ReadWrite"}}
                     },
                     None),
        },
        "test_dump_val": {
            "int": (b"\x01\x00\x00\x00", DeviceAttrTypeEnum.INT, "attr_int",
                    {"id": 1, "type": "int", "accessMode": "ReadWrite"}, 1),
            "float": (b"\x00\x00\x80?", DeviceAttrTypeEnum.FLOAT, "attr_float",
                      {"id": 2, "type": "float", "accessMode": "ReadWrite"}, 1.0),
            "longlong": (b"\x08\x00\x00\x00\x00\x00\x00\x00", DeviceAttrTypeEnum.LONG_LONG, "attr_longlong",
                         {"id": 3, "type": "long long", "accessMode": "ReadWrite"}, 8),
            "string": (b"\x61\x62\x63" + b"\x00" * (MAX_STRING_BUFFER_SIZE_BYTE - 3), DeviceAttrTypeEnum.STRING,
                       "attr_string", {"id": 4, "type": "string", "accessMode": "ReadWrite"}, "abc"),
            "bool": (b"\x01", DeviceAttrTypeEnum.BOOL, "attr_bool",
                     {"id": 5, "type": "bool", "accessMode": "ReadWrite"}, True),
            "json": (b"\x0a\x00\x00\x00\x01\x00\x00\x00\x01", DeviceAttrTypeEnum.JSON, "attr_json",
                     {
                         "id": 6, "type": "json", "accessMode": "ReadWrite",
                         "subAttributes": {"switch_on": {"id": 10, "type": "bool", "accessMode": "ReadWrite"}}
                     },
                     {"switch_on": True}),
        },
        "test_load_val": {
            "int": (1, DeviceAttrTypeEnum.INT, "attr_int",
                    {"id": 1, "type": "int", "accessMode": "ReadWrite"}, b"\x01\x00\x00\x00"),
            "float": (1.0, DeviceAttrTypeEnum.FLOAT, "attr_float",
                      {"id": 2, "type": "float", "accessMode": "ReadWrite"}, b"\x00\x00\x80?"),
            "longlong": (8, DeviceAttrTypeEnum.LONG_LONG, "attr_longlong",
                         {"id": 3, "type": "long long", "accessMode": "ReadWrite"},
                         b"\x08\x00\x00\x00\x00\x00\x00\x00"),
            "string": ("abc", DeviceAttrTypeEnum.STRING,
                       "attr_string", {"id": 4, "type": "string", "accessMode": "ReadWrite"},
                       b"\x61\x62\x63" + b"\x00" * (MAX_STRING_BUFFER_SIZE_BYTE - 3)),
            "bool": (True, DeviceAttrTypeEnum.BOOL, "attr_bool",
                     {"id": 5, "type": "bool", "accessMode": "ReadWrite"}, b"\x01"),
            "json": ({"switch_on": True}, DeviceAttrTypeEnum.JSON, "attr_json",
                     {
                         "id": 6, "type": "json", "accessMode": "ReadWrite",
                         "subAttributes": {"switch_on": {"id": 10, "type": "bool", "accessMode": "ReadWrite"}}
                     },
                     b"\x0a\x00\x00\x00\x01\x00\x00\x00\x01"),
        },
        "test_attr_factory_will_raise_exception_when_invalid_attr_type": {
            "unsigned int": (UnsupportedDeviceAttrTypeError, InvalidDeviceAttrTypeEnum.UINT, "attr_uint",
                             {"id": 1, "type": "unsigned int", "accessMode": "ReadWrite"}),
            "xml": (UnsupportedDeviceAttrTypeError, InvalidDeviceAttrTypeEnum.XML, "attr_xml",
                    {"id": 1, "type": "xml", "accessMode": "ReadWrite"}),
        },
        "test_attr_factory_will_ok_when_valid_attr_type": {
            "int": (DeviceAttrInt, DeviceAttrTypeEnum.INT, "attr_int",
                    {"id": 1, "type": "int", "accessMode": "ReadWrite"}),
            "json": (DeviceAttrJson, DeviceAttrTypeEnum.JSON, "attr_json",
                     {"id": 1, "type": "json", "accessMode": "ReadWrite"}),
        },
    }

    @staticmethod
    def test_load_tag_normal(model: LoadTagCase):
        assert DeviceAttr.load_tag(model.tag_buff) == model.expect

    @staticmethod
    def test_load_tag_error(model: LoadTagCase):
        with pytest.raises(model.expect):
            assert DeviceAttr.load_tag(model.tag_buff)

    @staticmethod
    def test_get_val_len(model: GetValueLenCase):
        attr_inst = attr_factory(model.attr_type, model.attr_name, model.attr_info)
        assert attr_inst.get_val_len(model.value) == model.expect

    @staticmethod
    def test_load_val(model: GetValueLenCase):
        attr_inst = attr_factory(model.attr_type, model.attr_name, model.attr_info)
        assert attr_inst.load_val(model.value) == model.expect

    @staticmethod
    def test_attr_factory_will_raise_exception_when_invalid_attr_type(model: AttrFactoryCase):
        with pytest.raises(model.expect):
            attr_factory(model.attr_type, model.attr_name, model.attr_info)

    @staticmethod
    def test_attr_factory_will_ok_when_valid_attr_type(model: AttrFactoryCase):
        attr_inst = attr_factory(model.attr_type, model.attr_name, model.attr_info)
        assert isinstance(attr_inst, model.expect)
