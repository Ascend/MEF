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
import json
import os
from typing import Dict

from common.constants.base_constants import CommonConstants
from common.constants.upgrade_constants import UpgradeConstants
from common.log.logger import run_log
from common.utils.exception_utils import OperateBaseError
from common.utils.singleton import Singleton
from common.utils.system_utils import SystemUtils
from common.utils.version_xml_file_manager import VersionXmlManager


class SettingInitError(OperateBaseError):
    pass


class SystemSetting(Singleton):

    def __init__(self):
        # A200产品不包含MCU与firmware固件
        self.board_type = self.get_board_type()

        if self.board_type == CommonConstants.ATLAS_200I_A2:
            self.custom_sdk_version = self.get_custom_sdk_version()

        if self.board_type == CommonConstants.ATLAS_500_A2:
            self.mcu_version = self.get_mcu_version()
            self.board_id = self.get_board_id()

        self.firmware_version = self.get_firmware_version()
        self.om_version = self.get_om_version()
        self.npu_version = self.get_npu_version()
        self.calculate_ability = self.get_npu_calculate_ability()

    @staticmethod
    def get_npu_calculate_ability() -> str:
        try:
            return SystemUtils.get_npu_calculate_ability_by_npu_smi()
        except Exception as err:
            run_log.error("failed to get npu calculate ability failed, catch %s", err)
            return "8TOPS"

    @staticmethod
    def get_board_type() -> str:
        """通过npu-smi info -t product -i 0从环境中获取板型号"""
        try:
            board_type = SystemUtils.get_model_by_npu_smi()
        except Exception as err:
            raise SettingInitError(f"Get model from by npu-smi err: {err}") from err

        if board_type in CommonConstants.A500_MODELS:
            return CommonConstants.ATLAS_500_A2

        elif board_type in CommonConstants.A200_MODELS:
            return CommonConstants.ATLAS_200I_A2

        raise SettingInitError("Get model from by npu-smi err, not support")

    @staticmethod
    def get_mcu_version() -> str:
        try:
            return SystemUtils.get_mcu_version_by_npu_smi()
        except Exception as err:
            run_log.error("failed to get mcu version failed, catch %s", err)
            return ""

    @staticmethod
    def get_board_id() -> str:
        try:
            return SystemUtils.get_board_id_by_npu_smi()
        except Exception as err:
            run_log.error("failed to get board id failed, catch %s", err)
            return ""

    @staticmethod
    def get_firmware_version() -> str:
        try:
            return VersionXmlManager(CommonConstants.OM_VERSION_XML_FILE).firmware_version
        except Exception as err:
            run_log.error("failed to get firmware version. catch %s", err)
            return ""

    @staticmethod
    def get_om_version() -> str:
        try:
            return VersionXmlManager(CommonConstants.OM_VERSION_XML_FILE).version
        except Exception as err:
            run_log.error("failed to get om version. catch %s", err)
            return ""

    @staticmethod
    def get_custom_sdk_version() -> Dict[str, str]:
        if not os.path.exists(UpgradeConstants.SDK_PACKAGE_VERSION_JSON):
            return {}
        try:
            with open(UpgradeConstants.SDK_PACKAGE_VERSION_JSON, "r") as file:
                return json.loads(file.read())
        except Exception as err:
            run_log.error("failed to get custom sdk version. catch %s", err)
            return {}

    @staticmethod
    def get_npu_version() -> str:
        try:
            return SystemUtils.get_npu_version_by_npu_smi()
        except Exception as err:
            run_log.error("failed to get npu version failed, catch %s", err)
            return ""


class MonitorBackupRestoreCfg:
    """ monitor进程备份恢复配置 """

    BACKUP_FILES = (
        "/home/data/ies/access_control.ini",
        "/home/data/ies/certWarnTime.ini",
        "/home/data/ies/ibma_edge_service.ini",
        "/home/data/ies/mountCnf_site.ini",
        "/home/data/ies/shield_alarm.ini",
        "/home/data/ies/NTPEnable.ini",
        "/home/data/config/monitor/monitor_edge.db",
    )
    BACKUP_DIR: str = CommonConstants.MONITOR_BACKUP_DIR
    DELETE_BLACKLIST = (
        "/home/data/ies/mountCnf_site.ini",
        "/home/data/ies/shield_alarm.ini",
        "/home/data/ies/NTPEnable.ini",
        "/home/data/config/monitor/monitor_edge.db",
    )
