# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from dataclasses import dataclass
from typing import List, Dict, Optional

from common.schema import BaseModel, field


@dataclass
class Firmware(BaseModel):
    """上报给FD的firmware_list中的记录结构"""
    name: str = field(alias="Module", comment="固件包名")
    version: str = field(alias="Version", comment="当前运行版本号")
    inactive_version: str = field(default="", alias="InactiveVersion", comment="待生效版本号")
    active_method: str = field(default="inband", comment="生效方式，似乎冗余字段")
    board_id: str = field(default="", alias="BoardId", comment="FD对比用")
    upgrade_agent: str = field(default="OM", comment="升级代理，FD判断用")
    upgrade_result: Optional[bool] = field(default=None, alias="UpgradeResult", comment="升级状态，to_dict时去掉")

    def __post_init__(self):
        # 未升级或升级失败，上报给FD的待生效版本号为空
        if not self.upgrade_result:
            self.inactive_version = ""

    @classmethod
    def firmware_list(cls, data: List[Dict[str, str]]) -> List[Dict[str, str]]:
        """将底层systems传过来的数据转换成FD需要的数据结构。执行到此，已用模板校验过结构，因此结构一定满足"""
        return [cls.from_dict(ds).to_dict() for ds in data]

    def to_dict(self):
        data = super().to_dict()
        data.pop("upgrade_result")
        return data
