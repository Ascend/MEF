# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from dataclasses import dataclass
from typing import Optional, Callable, Dict

from bin.monitor_config import SystemSetting
from common.constants.base_constants import CommonConstants
from common.constants.upgrade_constants import UpgradeConstants
from common.schema import field
from lib.Linux.upgrade.upgrade_new import Upgrade


@dataclass
class FirmwareInfo:
    """固件信息"""
    name: str = field(comment="固件包名, ModuleType中的一个")
    version: str = field(default="", comment="当前运行版本号")
    inactive_version: str = field(default="", comment="待生效版本号")
    board_id: str = field(default="", comment="运行版本board_id由驱动接口获取")
    upgrade_result: Optional[bool] = field(default=None, comment="升级结果, None表示未升级")
    upgrade_process: int = field(default=0, comment="升级进度")

    def __post_init__(self):
        # 获取固件运行版本的版本
        ModuleVersionManager.set_cur_version_info(self)

    def get_module_msg(self):
        """将实例转换成字典，给Systems用,需要与systems模板文件中Firmware嵌套结构中的字段对应"""
        module = Upgrade.get_modules().get(self.name)
        if module:
            self.inactive_version = module.version
            self.upgrade_result = module.state
            self.upgrade_process = module.process

        return {
            "Version": self.version,
            "InactiveVersion": self.inactive_version,
            "Module": self.name,
            "BoardId": self.board_id,
            "UpgradeResult": self.upgrade_result,
            "UpgradeProcess": self.upgrade_process,
        }


class ModuleVersionManager:

    @staticmethod
    def get_firmware_info(module: FirmwareInfo):
        """获取firmware当前运行版本号,其他信息给默认值"""
        # 当板型号为A500时，Firmware版本为固件版本。当板型号为A200时，Firmware版本即为OM版本
        sets = SystemSetting()
        module.version = sets.om_version if sets.board_type == CommonConstants.ATLAS_200I_A2 else sets.firmware_version

    @staticmethod
    def get_mcu_info(module: FirmwareInfo):
        """获取mcu的当前运行版本信息"""
        # A200 不存在MCU芯片
        if SystemSetting().board_type == CommonConstants.ATLAS_200I_A2:
            return
        module.version = SystemSetting().mcu_version
        module.board_id = SystemSetting().board_id

    @staticmethod
    def get_npu_info(module: FirmwareInfo):
        """获取npu的当前运行版本信息，只获取版本号"""
        module.version = SystemSetting().npu_version

    @staticmethod
    def set_cur_version_info(module: FirmwareInfo):
        """获取当前运行版本的版本号"""
        set_info_func_map: Dict[str, Callable] = {
            UpgradeConstants.FIRMWARE_TYPE: ModuleVersionManager.get_firmware_info,
            UpgradeConstants.NPU_TYPE: ModuleVersionManager.get_npu_info,
            UpgradeConstants.MCU_TYPE: ModuleVersionManager.get_mcu_info,
            UpgradeConstants.OMSDK_TYPE: ModuleVersionManager.get_firmware_info
        }
        set_info_func_map.get(module.name, ModuleVersionManager.get_custom_sdk_version_info)(module)

    @staticmethod
    def get_custom_sdk_version_info(module: FirmwareInfo):
        """第三方固件版本从数据库获取"""
        module.version = SystemSetting().custom_sdk_version.get(module.name, "")
