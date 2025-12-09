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
from enum import Enum
from typing import Dict
from typing import NewType
from typing import Union

DevmValType = NewType("DevmValType", Union[int, float, str, Dict])


class ModuleCategoryEnum(Enum):
    """模组类型枚举"""

    INTERNAL = "internal"
    EXTEND = "extend"
    ADDITION = "addition"


class DeviceAttrAccessModeEnum(Enum):
    """设备属性访问模式枚举"""

    R = "Read"
    W = "Write"
    RW = "ReadWrite"
    RH = "ReadHide"
    WH = "WriteHide"
    RWH = "ReadWriteHide"

    @property
    def is_readable(self):
        return self == self.R or self == self.RW or self == self.RH or self == self.RWH

    @property
    def is_writeable(self):
        return self == self.W or self == self.RW or self == self.WH or self == self.RWH


class DeviceAttrTypeEnum(Enum):
    """设备属性类型枚举"""

    INT = "int"
    FLOAT = "float"
    BOOL = "bool"
    LONG_LONG = "long long"
    STRING = "string"
    JSON = "json"
