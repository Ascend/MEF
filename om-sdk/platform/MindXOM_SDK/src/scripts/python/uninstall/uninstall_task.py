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
import signal
import sys
from itertools import chain
from typing import Generator, Callable, Any, Tuple

from common.constants.base_constants import CommonConstants
from common.constants.upgrade_constants import OMUpgradeConstants
from common.file_utils import FileAttribute, FileUtils
from common.file_utils import FileCheck
from common.utils.exception_utils import OperateBaseError
from common.utils.exec_cmd import ExecCmd
from common.utils.scripts_utils import signal_handler
from lib.Linux.systems.nfs.nfs_manage import NfsManage
from logger import upgrade_log, terminal_print
from uninstall.task import BaseTask
from utils import get_login_user, OperationRetCode


class UninstallError(OperateBaseError):
    pass


class UnsetImmutableAttr(BaseTask):
    name = "Unset work dir immutable attribute task"

    @staticmethod
    def _check_work_dir():
        if not FileCheck.check_path_is_exist_and_valid(CommonConstants.OM_WORK_DIR_PATH):
            # 工作目录不存在或异常时，卸载失败
            raise UninstallError("MindXOM work dir invalid, uninstall failed.")

    @staticmethod
    def _unset_immutable_attr():
        ret = FileAttribute.set_path_immutable_attr(CommonConstants.OM_WORK_DIR_PATH, False)
        if not ret:
            raise UninstallError(ret.error)

    def steps(self) -> Generator[Callable[..., Any], None, None]:
        yield self._check_work_dir
        yield self._unset_immutable_attr


class RemoveService(BaseTask):
    SERVICES: Tuple[str] = ("om-init", "platform-app", "start-nginx", "ibma-edge-start",)
    SERVICE_DIR: str = "/usr/lib/systemd/system"
    SERVICE_LINK_DIR: str = "/etc/systemd/system/multi-user.target.wants"

    name = "Remove service task"

    @staticmethod
    def _daemon_reload():
        ret = ExecCmd.exec_cmd(("systemctl", "daemon-reload"))
        if ret != 0:
            upgrade_log.warning("daemon reload error: %s", ret)

    def steps(self) -> Generator[Callable[..., Any], None, None]:
        yield self._stop_service
        yield self._remove_service_link
        yield self._remove_services_and_scripts
        yield self._daemon_reload

    def _stop_service(self):
        for server in self.SERVICES:
            ret = ExecCmd.exec_cmd(("systemctl", "stop", f"{server}"))
            if ret != 0:
                # 停服务失败，触发卸载失败
                raise UninstallError(f"Stop {server} failed, return code: {ret}")

    def _remove_service_link(self):
        for server in self.SERVICES:
            link_file = os.path.join(self.SERVICE_LINK_DIR, f"{server}.service")
            try:
                FileUtils.delete_file_or_link(link_file)
            except Exception as err:
                # 文件删除失败，只是warn,产生残留
                upgrade_log.warning("Remove %s failed, catch %s. Please clean manually.",
                                    link_file, err.__class__.__name__)

    def _remove_services_and_scripts(self):
        services = (os.path.join(self.SERVICE_DIR, f"{server}.service") for server in self.SERVICES)
        for filename in services:
            try:
                FileUtils.delete_file_or_link(filename)
            except Exception as err:
                # 文件删除失败，只是warn,产生残留
                upgrade_log.warning("Remove %s failed, catch %s. Please clean manually.",
                                    filename, err.__class__.__name__)


class ClearConfig(BaseTask):
    IES_DIR: str = "/home/data/ies"
    IES_CONFIGS: Tuple[str] = (
        "ibma_edge_service.ini", "tag.ini", "ies_flag", "NTPEnable.ini",
        "certWarnTime.ini", "shield_alarm.ini", "mountCnf_site.ini", "access_control.ini"
    )
    CONFIG_DIRS: Tuple[str] = ("/home/data/config", "/home/data/etc/root", "/run/web", "/run/upgrade", "/run/nfs")
    OTHERS: Tuple[str] = (
        OMUpgradeConstants.IMPORT_CERT_FLAG,
        "/run/certWarnTimeUpdate",
        OMUpgradeConstants.UPGRADE_FINISHED_FLAG,
        OMUpgradeConstants.OM_RESET_FLAG,
        CommonConstants.IBMA_SOCK_PATH,
        "/run/nfs/nfs_status_info",
    )

    name = "Clear config task"

    def steps(self) -> Generator[Callable[..., Any], None, None]:
        yield NfsManage().uninstall
        yield self._remove_config_files
        yield self._remove_config_dir

    def _ies_config_files(self) -> Generator[str, None, None]:
        for filename in self.IES_CONFIGS:
            yield os.path.join(self.IES_DIR, filename)

    def _remove_config_dir(self):
        for cfg_dir in self.CONFIG_DIRS:
            if not FileUtils.delete_dir_content(cfg_dir):
                upgrade_log.warning("Remove config dir '%s' failed, not exists or invalid.", os.path.basename(cfg_dir))

    def _remove_config_files(self):
        for filename in chain(self._ies_config_files(), self.OTHERS):
            try:
                FileUtils.delete_file_or_link(filename)
            except Exception as err:
                # 文件删除失败，只是warn,产生残留
                upgrade_log.warning("Remove %s failed, catch %s. Please clean manually.", filename, err)


def uninstall() -> int:
    upgrade_log.info("Start uninstall ...")

    # 获取当前操作的用户与ip，校验当前操作的用户是否为root
    op_user, _ = get_login_user()
    if op_user != "root":
        upgrade_log.error("Current user %s is invalid, please use root to uninstall MindXOM.", op_user)
        return OperationRetCode.FAILED_OPERATION

    # 开始执行卸载
    for task in UnsetImmutableAttr, RemoveService, ClearConfig:
        upgrade_log.info("%s start.", task.name)
        terminal_print.info("%s start.", task.name)
        try:
            task().run()
        except UninstallError as err:
            upgrade_log.error("%s failed, because of %s", task.name, err)
            return OperationRetCode.FAILED_OPERATION

        upgrade_log.info("%s success", task.name)
        terminal_print.info("%s success", task.name)

    return OperationRetCode.SUCCESS_OPERATION


if __name__ == '__main__':
    # 注册退出信号的中断处理函数
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    try:
        sys.exit(uninstall())
    except Exception as error:
        upgrade_log.error("uninstall MindXOM failed. Catch %s", error.__class__.__name__)
        sys.exit(1)
