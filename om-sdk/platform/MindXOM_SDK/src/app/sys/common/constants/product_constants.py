# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os

from common.ResourceDefV1.service_root import RfServiceRoot
from common.utils.app_common_method import AppCommonMethod

# 重启方式
REDFISH_RESTART_TYPE = ("GracefulRestart",)
FD_RESTART_TYPE = ("Graceful",)

# 日志收集范围
LOG_COLLECT_LIST = ["NPU", "MindXOM", "MEF", "OSDivers"]
LOG_MODULES_MAP = {"all": "NPU MindXOM MEF OSDivers"}
COLLECT_LOG_SHELL_PATH = "lib/Linux/systems/disk/log_collect.sh"

# redfish响应根节点实例
PRJ_DIR: str = AppCommonMethod.get_project_absolute_path()
MOCKUP_PATH: str = os.path.join(PRJ_DIR, "common/MockupData/iBMAServerV1")
ROOT_PATH: str = os.path.normpath("redfish/v1")
SERVICE_ROOT = RfServiceRoot(MOCKUP_PATH, ROOT_PATH)

# systems产品信息--产品型号
MODEL = "Atlas 200 A2"


# 下载接口
DOWNLOAD_FUNC: tuple = (
    "systems.rf_system_log_download",
    "systems.rf_export_system_puny_dict",
    "systems.rf_download_csr_file",
    "systems.rf_export_security_load",
)


