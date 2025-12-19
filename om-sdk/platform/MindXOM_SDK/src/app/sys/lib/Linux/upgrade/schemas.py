# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from dataclasses import dataclass, fields
from typing import Optional, Set

from common.schema import field


@dataclass
class ModuleState:
    version: str = field(default="", comment="升级版本号，对应version.xml中的Version")
    name: str = field(default="", comment="升级版模块名，对应version.xml中的Module")
    state: Optional[bool] = field(default=None, comment="升级成功状态，None表示未进行过升级")
    process: int = field(default=0, comment="升级进度")

    def update_front_msg(self, other: "ModuleState"):
        """更新返回给前端的升级信息"""
        for _field in fields(self):
            self.__setattr__(_field.name, other.__getattribute__(_field.name))
