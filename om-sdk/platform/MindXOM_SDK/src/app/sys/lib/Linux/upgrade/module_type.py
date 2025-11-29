# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from enum import Enum
from typing import Set

from bin.monitor_config import SystemSetting
from common.constants.base_constants import CommonConstants
from common.constants.upgrade_constants import UpgradeConstants


class A500ModuleType(Enum):
    FIRMWARE = UpgradeConstants.FIRMWARE_TYPE
    MCU = UpgradeConstants.MCU_TYPE
    NPU = UpgradeConstants.NPU_TYPE

    @classmethod
    def values(cls) -> Set[str]:
        return {elem.value for elem in cls}


class A200ModuleType(Enum):
    FIRMWARE = UpgradeConstants.OMSDK_TYPE

    @classmethod
    def values(cls) -> Set[str]:
        return {elem.value for elem in cls}


ModuleType = A500ModuleType if SystemSetting().board_type == CommonConstants.ATLAS_500_A2 else A200ModuleType
