# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os
import threading
import time

from common.constants.product_constants import REDFISH_RESTART_TYPE
from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd
from devm.device_mgr import DEVM
from devm.exception import DeviceManagerError, MethodNotAllowedError
from lib.Linux.systems.actions import SystemAction
from lib.Linux.upgrade.upgrade_effect import UpgradeEffect
from lib.Linux.upgrade.upgrade_new import Upgrade
from common.common_methods import CommonMethods


class OMSystemAction(SystemAction):
    """
    功能描述：系统action
    接口：NA
    """
    @classmethod
    def force_restart(cls):
        """
        注意：此操作一定会成功。
        """
        try:
            ExecCmd.exec_cmd(cls.LOG_SAVE_CMD)
            os.sync()
            time.sleep(cls.RESTART_DELAY_TIME)
            ExecCmd.exec_cmd(cls.FORCE_RESTART_CMD)
        except Exception as err:
            run_log.error("reboot system failed, because unknown error %s", err)
        finally:
            cls.EDGE_SYSTEM_ACTION_LOCK.release()

    @classmethod
    def cold_restart(cls, instance, attr):
        """
        注意：此操作成功与否由驱动把控，如果此处抛出异常一定是驱动的问题，对于OM来说走到此函数了一定会响应成功
        """
        try:
            ExecCmd.exec_cmd(cls.LOG_SAVE_CMD)
            os.sync()
            time.sleep(cls.RESTART_DELAY_TIME)
            instance.set_attr_by_module(attr, 2)
        except Exception as err:
            run_log.error("reboot system failed, because unknown error %s", err)
        finally:
            cls.EDGE_SYSTEM_ACTION_LOCK.release()

    @staticmethod
    def get_device_instance():
        instance = DEVM.get_device("system0")
        attr = instance.get_attr_by_module("cold_reset_system")
        if not attr.access_mode.is_writeable:
            raise MethodNotAllowedError("attr access mode is not writeable")

        return instance, attr

    @staticmethod
    def post_request(request_data_dict):
        if OMSystemAction.EDGE_SYSTEM_ACTION_LOCK.locked():
            run_log.error("The operation is busy")
            return [CommonMethods.ERROR, "Computer System Reset fail."]
        OMSystemAction.EDGE_SYSTEM_ACTION_LOCK.acquire(OMSystemAction.LOCK_TIMEOUT)

        action = request_data_dict.get("ResetType", None)
        if action not in REDFISH_RESTART_TYPE:
            OMSystemAction.EDGE_SYSTEM_ACTION_LOCK.release()
            return [CommonMethods.ERROR, "Computer System Reset fail."]

        if Upgrade.allow_effect():
            try:
                UpgradeEffect().effect_firmware()
            except Exception as err:
                run_log.error("restart to effect failed: %s", err)
                OMSystemAction.EDGE_SYSTEM_ACTION_LOCK.release()
                return [CommonMethods.ERROR, "Restart to effect failed system failed."]

        run_log.info("Start to reset system by %s.", action)
        if action == "ForceRestart":  # 强制重启
            threading.Thread(target=OMSystemAction.force_restart).start()
            return [CommonMethods.OK, f"Restart system ({action}) successfully."]
        elif action == "GracefulRestart":  # 优雅重启
            threading.Thread(target=OMSystemAction.graceful_restart).start()
            return [CommonMethods.OK, f"Restart system ({action}) successfully."]
        elif action == "Restart":  # 冷复位
            try:
                instance, attr = OMSystemAction.get_device_instance()
            except DeviceManagerError as err:
                run_log.error("Get device instance failed, %s.", err)
                OMSystemAction.EDGE_SYSTEM_ACTION_LOCK.release()
                return [CommonMethods.ERROR, f"Restart system ({action}) failed."]

            threading.Thread(target=OMSystemAction.cold_restart, args=(instance, attr)).start()
            return [CommonMethods.OK, f"Restart system ({action}) successfully."]

        return [CommonMethods.OK, ]
