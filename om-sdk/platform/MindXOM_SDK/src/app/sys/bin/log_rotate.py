# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os
import shlex
import tarfile
import time
from pathlib import Path
from typing import List
from typing import NoReturn

from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.file_utils import FileCopy
from common.file_utils import FileCreate
from common.file_utils import FilePermission as Chmod
from common.file_utils import FileUtils
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.exception_utils import OperateBaseError
from common.utils.exec_cmd import ExecCmd
from common.utils.system_utils import SystemUtils


class LogRotateError(OperateBaseError):
    pass


class LogRotate:
    FULL_CHECK_PERIOD = 600
    ROTATE_CHECK_TIMES = 6
    CMD_WAIT_TIME = 60
    VAR_PATH = "/var"
    VAR_PLOG_PATH = os.path.join(VAR_PATH, "plog")
    HOME_LOG_PATH = "/home/log"
    HOME_LOG_PLOG_PATH = os.path.join(HOME_LOG_PATH, "plog")
    LOG_SYNC_LOG_PATH = os.path.join(HOME_LOG_PATH, "om_sync")
    LOGROTATE_CONF = os.path.join(CommonConstants.OM_WORK_DIR_PATH, "config", "plog")
    INVALID_FILE_SIZE = 150  # 150M
    B_TO_MB = 1024 * 1024
    LOG_COLLECTIONS = (
        ("upgrade.log", "", CommonConstants.MONITOR_USER),
        ("ntp_service.log", "", CommonConstants.MONITOR_USER),
        ("om_platform_run.log", "ibma_edge", CommonConstants.MONITOR_USER),
        ("om_scripts_run.log", "ibma_edge", CommonConstants.MONITOR_USER),
        ("seceye_agent_run.log", "ibma_edge", CommonConstants.MONITOR_USER),
        ("manager_run.log", "manager", CommonConstants.MONITOR_USER),
        ("manager_operate.log", "manager", CommonConstants.MONITOR_USER),
        ("redfish_operate.log", "redfish", CommonConstants.MINDXOM_USER),
        ("redfish_run.log", "redfish", CommonConstants.MINDXOM_USER),
        ("redfish_platform_run.log", "redfish", CommonConstants.MINDXOM_USER),
        ("access.log", "web_edge", CommonConstants.NGINX_USER),
        ("error.log", "web_edge", CommonConstants.NGINX_USER),
        ("web_edge_run.log", "web_edge", CommonConstants.NGINX_USER),
    )
    RUN_PATH = "/run"
    DISK_MAX_SIZE = 300  # 单位是MB
    RAM_MAX_SIZE = 100
    MAX_LOOP_NUM = 1000

    def start_logrotate(self) -> NoReturn:
        run_log.info("logrotate service starting.....")
        try:
            self._log_file_check()
        except Exception as err:
            run_log.error("logrotate failed and restart, because %s", err)
            self.start_logrotate()

    def _log_sync(self) -> NoReturn:
        """对运行日志进行同步，防止掉电丢失"""
        if not SystemUtils.is_tmpfs_type(self.VAR_PLOG_PATH):
            return

        if not os.path.exists(self.LOG_SYNC_LOG_PATH):
            if not FileCreate.create_dir(self.LOG_SYNC_LOG_PATH, 0o750):
                raise LogRotateError(f"create log sync dir {self.LOG_SYNC_LOG_PATH} failed")

        for log_name, dir_name, user in self.LOG_COLLECTIONS:
            # 创建同步目录
            src_log = os.path.join(self.VAR_PLOG_PATH, dir_name, log_name)
            if not FileCheck.check_path_is_exist_and_valid(src_log):
                run_log.error("log file [%s] not valid", src_log)
                continue
            dst_dir = os.path.join(self.LOG_SYNC_LOG_PATH, dir_name)
            if not os.path.exists(dst_dir):
                if not FileCreate.create_dir(dst_dir, 0o750):
                    run_log.error("create log sync dir %s failed", dst_dir)
                    continue
                if not Chmod.set_path_owner_group(dst_dir, user):
                    run_log.error("set owner log sync dir %s failed", dst_dir)
                    continue
            # 日志同步
            dst_log = os.path.join(dst_dir, log_name)
            sync_cmd = ("rsync", "--inplace", "-az", src_log, dst_log)
            if ExecCmd.exec_cmd(sync_cmd) != 0:
                run_log.error("rsync om log %s failed", log_name)
                continue

            run_log.info("rsync om log %s success", log_name)

    def _home_log_check_and_storage(self, check_dir: Path, check_size: int) -> NoReturn:
        """空间满，直接删除最老的"""
        used_size = SystemUtils.get_fs_used_size(check_dir)
        log_list = self._get_all_file_list(check_dir)
        log_list.sort(key=lambda x: os.path.getsize(x))
        while log_list and used_size > check_size:
            log_file = log_list.pop()
            # 如果文件大小超过150M，认为是非法文件，先删除
            file_size = os.path.getsize(log_file) / self.B_TO_MB
            if file_size > self.INVALID_FILE_SIZE:
                FileUtils.delete_file_or_link(log_file)
                used_size -= file_size
            else:
                log_list.append(log_file)
                break

        # 按照修改日期排序，排在最后面的是修改日期最久的文件
        log_list.sort(key=lambda x: os.path.getmtime(x), reverse=True)
        while log_list and used_size >= check_size:
            # 删除修改日期最久的文件
            log_file = log_list.pop()
            used_size -= os.path.getsize(log_file) / self.B_TO_MB
            FileUtils.delete_file_or_link(log_file)

        run_log.info("log disk used size check finished.")

    def _var_log_check_and_storage(self, log_dir: Path, max_size: int, storage_dir: Path) -> NoReturn:
        """空间满，将最大的文件转存,以M为单位"""
        if not storage_dir.exists():
            run_log.info("The log %s is not exist, creat it.", storage_dir)
            storage_dir.mkdir(mode=0o750)

        used_size = SystemUtils.get_fs_used_size(log_dir)
        log_list = self._get_all_file_list(log_dir)
        # 按照文件大小排序，排在最后面的是最大的文件
        log_list.sort(key=lambda x: os.path.getsize(x))
        while log_list and used_size >= max_size:
            # 将最大的文件备份一份，用以转储
            largest_file = log_list.pop()
            u_name = largest_file.owner()
            largest_file_back = os.path.join(self.RUN_PATH, f"{largest_file.name}.1")
            ret = FileCopy.copy_file(largest_file.as_posix(), largest_file_back)
            if not ret:
                raise LogRotateError(f"Copy {largest_file.name} failed, because {ret.error}")

            # 清空最大的文件的内容
            cmd = (
                "su", "-", u_name, "-s", cmd_constants.OS_CMD_BASH, "-c",
                f"{cmd_constants.OS_CMD_CAT} /dev/null > {shlex.quote(largest_file.as_posix())}"
            )
            ret = ExecCmd.exec_cmd_get_output(cmd, wait=self.CMD_WAIT_TIME)
            if ret[0] != 0:
                FileUtils.delete_file_or_link(largest_file_back)
                raise LogRotateError(f"exec cmd failed.")

            # 将日志压缩转储到目标目录，响应日志放到相对应目录下
            now_time = time.strftime("%Y%m%d_%H%M%S")
            dest_tar_dir = storage_dir.parent.joinpath(largest_file.parent.relative_to(self.VAR_PATH))
            if not dest_tar_dir.exists():
                dest_tar_dir.mkdir(mode=0o750)
            dest_tar_path = dest_tar_dir.joinpath(f"{largest_file.stem}_{now_time}.tar.gz")
            try:
                with tarfile.open(dest_tar_path, "w:gz") as tar:
                    tar.add(largest_file_back)
            except Exception as err:
                FileUtils.delete_file_or_link(largest_file_back)
                raise LogRotateError(f"Failed to package log file, because unknown error {err}") from err

            # 删除/run目录下的临时备份文件
            try:
                FileUtils.delete_file_or_link(largest_file_back)
            except Exception as err:
                raise LogRotateError(f"delete {os.path.basename(largest_file_back)} failed, because {err}") from err

            # 更新当前使用大小
            used_size = SystemUtils.get_fs_used_size(log_dir)

        run_log.info("log ram used ratio check finished.")

    def _log_file_check(self) -> NoReturn:
        """每10分钟检测一次空间，进行一次日志同步，与日志目录空间检测"""
        count = 0
        while True:
            # 每隔1小时通过 logrotate 对/var/plog目录进行日志转储
            count += 1
            if count >= self.ROTATE_CHECK_TIMES:
                count = 0
                self._var_plog_logrotate()

            # 日志同步
            self._log_sync()

            # 空间满检测并转储
            self._var_log_check_and_storage(Path(self.VAR_PLOG_PATH), self.RAM_MAX_SIZE, Path(self.HOME_LOG_PLOG_PATH))
            self._home_log_check_and_storage(Path(self.HOME_LOG_PATH), self.DISK_MAX_SIZE)

            # 将所有文件修改为440权限，目录修改为750权限
            ret = Chmod.set_path_permission(self.HOME_LOG_PLOG_PATH, 0o440, recursive=True, ignore_file=False)
            if not ret:
                raise LogRotateError(f"chmod failed, because {ret.error}")
            ret = Chmod.set_path_permission(self.HOME_LOG_PLOG_PATH, 0o750, recursive=True, ignore_file=True)
            if not ret:
                raise LogRotateError(f"chmod failed, because {ret.error}")

            time.sleep(self.FULL_CHECK_PERIOD)

    def _get_all_file_list(self, path: Path) -> List[Path]:
        log_list = []
        # rglob("*") 返回当前目录及其下所有子目录中的【所有文件夹】和【所有文件】
        for file in path.rglob("*"):
            if len(log_list) > self.MAX_LOOP_NUM:
                run_log.warning("The number of log files exceeds the maximum %s limit", self.MAX_LOOP_NUM)
                return log_list

            if file.is_file():
                log_list.append(file)
        return log_list

    def _var_plog_logrotate(self) -> NoReturn:
        cmd = (cmd_constants.OS_CMD_LOGROTATE, self.LOGROTATE_CONF)
        ret = ExecCmd.exec_cmd_get_output(cmd, wait=self.CMD_WAIT_TIME)
        if ret[0] != 0:
            raise LogRotateError(f"logrotate execute failed, {ret[1]}")
        run_log.info("logrotate plog success")
