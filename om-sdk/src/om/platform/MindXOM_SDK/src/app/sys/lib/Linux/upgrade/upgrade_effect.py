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
import json
import os
import threading
import time
from functools import partial
from typing import NoReturn, List, Union

from bin.global_exclusive_control import GlobalExclusiveController
from bin.monitor_config import SystemSetting
from common.constants.base_constants import CommonConstants
from common.constants.upgrade_constants import (UpgradeEffectConstants, EffectResult, OMUpgradeConstants,
                                                UpgradeConstants)
from common.file_utils import FileUtils
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.exception_utils import OperateBaseError
from common.utils.exec_cmd import ExecCmd
from lib.Linux.upgrade.upgrade_new import Upgrade
from common.common_methods import CommonMethods


class EffectError(OperateBaseError):
    pass


class UpgradeEffect:

    LOCK_TIMEOUT = 120
    effect_lock = GlobalExclusiveController()
    exec_cmd = partial(ExecCmd.exec_cmd)

    @staticmethod
    def update_custom_version() -> NoReturn:
        version_dict = {}
        # 获取升级成功的组件版本
        for module in Upgrade.get_modules().values():
            if module.version and module.state:
                version_dict[module.name] = module.version

        # 更新json配置
        cur_version_msg = version_dict
        if os.path.exists(UpgradeConstants.SDK_PACKAGE_VERSION_JSON):
            with open(UpgradeConstants.SDK_PACKAGE_VERSION_JSON, "r") as file:
                cur_version_msg = json.loads(file.read())
            cur_version_msg.update(version_dict)

        flags = os.O_WRONLY | os.O_CREAT | os.O_TRUNC
        with os.fdopen(os.open(UpgradeConstants.SDK_PACKAGE_VERSION_JSON, flags, 0o600), "w") as file:
            file.write(json.dumps(cur_version_msg))

    def effect(self) -> NoReturn:
        time.sleep(UpgradeEffectConstants.RESTART_DELAY_TIME)
        run_log.info("Restart to effect firmware.")
        try:
            # 重启生效
            self.exec_cmd((cmd_constants.OS_CMD_SYSTEMCTL, "reboot"))
        except Exception as err:
            run_log.error("upgrade effect failed, because unknown error %s", err)
        finally:
            self.effect_lock.release()

    def effect_firmware(self):
        if SystemSetting().board_type == CommonConstants.ATLAS_500_A2:
            return

        # 更新用户自定义固件版本信息
        try:
            self.update_custom_version()
        except Exception as err:
            raise EffectError("update custom version failed") from err
        run_log.info("update custom version success")

        # OMSDK重启生效前，需要注册系统服务
        if not Upgrade.omsdk_upgraded():
            return

        run_log.info("copy MindXOM files to system dir.")
        if not FileUtils.check_script_file_valid(OMUpgradeConstants.OM_EFFECT_SHELL, "root", "root"):
            raise EffectError("copy_om_sys_file.sh invalid")

        if ExecCmd.exec_cmd((
                OMUpgradeConstants.OM_EFFECT_SHELL, UpgradeConstants.OMSDK_TYPE,
                OMUpgradeConstants.COPY_UPGRADE, CommonConstants.OM_UPGRADE_DIR_PATH
        )) != 0:
            raise EffectError("copy MindXOM files to system path failed.")

        # 保证数据落盘
        os.sync()

    def post_request(self, request_data: dict) -> List[Union[int, str]]:
        if self.effect_lock.locked():
            run_log.error("The operation is busy.")
            ret_msg = EffectResult.EFFECT_CODE_MAP.get(EffectResult.ERR_EFFECT_BUSY, "Unknown result")
            return [CommonMethods.ERROR, ret_msg]

        self.effect_lock.acquire(self.LOCK_TIMEOUT)

        # 参数是否合法
        if not isinstance(request_data, dict) or request_data.get("ResetType") not in ("GracefulRestart",):
            ret_msg = EffectResult.EFFECT_CODE_MAP.get(EffectResult.ERR_INVALID_PARAM, "Unknown result")
            self.effect_lock.release()
            return [CommonMethods.ERROR, ret_msg]

        # 是否允许生效
        if not Upgrade.allow_effect():
            ret_msg = EffectResult.EFFECT_CODE_MAP.get(EffectResult.ERR_UPGRADE_FAILED, "Unknown result")
            self.effect_lock.release()
            return [CommonMethods.ERROR, ret_msg]

        try:
            self.effect_firmware()
        except Exception as err:
            self.effect_lock.release()
            return [CommonMethods.ERROR, err.err_msg if isinstance(err, EffectError) else "Unknown result"]

        # 保证redfish可以记录操作日志并返回
        threading.Thread(target=self.effect).start()
        ret_msg = EffectResult.EFFECT_CODE_MAP.get(EffectResult.ERR_EFFECT_SUCCEED, "Unknown result")
        return [CommonMethods.OK, ret_msg]
