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
