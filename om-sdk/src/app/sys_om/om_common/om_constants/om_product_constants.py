# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os

from common.utils.app_common_method import AppCommonMethod
from om_common.ResourceDefV1.om_service_root import OMRfServiceRoot

# 重启方式
REDFISH_RESTART_TYPE = ("GracefulRestart", "ForceRestart", "Restart")
FD_RESTART_TYPE = ("Graceful", "Force", "ColdReset")

# 日志收集范围
LOG_COLLECT_LIST = ["NPU", "MindXOM", "MEF", "OSDivers", "MCU"]
LOG_MODULES_MAP = {"all": "NPU MCU MindXOM MEF OSDivers"}
COLLECT_LOG_SHELL_PATH = "om_lib/Linux/om_systems/disk/om_log_collect.sh"

# redfish响应根节点实例
PRJ_DIR: str = AppCommonMethod.get_project_absolute_path()
MOCKUP_PATH: str = os.path.join(PRJ_DIR, "common/MockupData/iBMAServerV1")
ROOT_PATH: str = os.path.normpath("redfish/v1")
SERVICE_ROOT = OMRfServiceRoot(MOCKUP_PATH, ROOT_PATH)

# systems产品信息--产品型号
MODEL = "Atlas 500 A2"

# 下载接口
DOWNLOAD_FUNC: tuple = ("systems.rf_system_log_download", "systems.rf_export_system_puny_dict",
                        "systems.rf_download_csr_file", "systems.rf_export_security_load",)
