#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
"""
用于Redfish Schema文件资源扩展
"""
import enum


class ExtendSchemaCollection(enum.Enum):
    """ 当前系统扩展所用Schema文件集合 """
    SCHEMA_EVENTSERVICE = "EventService.v1_8_0"

    @classmethod
    def get_extend_schema_list(cls):
        return [member.value for member in cls]
