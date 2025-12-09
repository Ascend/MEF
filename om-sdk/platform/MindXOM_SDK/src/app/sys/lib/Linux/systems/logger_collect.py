#!/usr/bin/python
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
import os
import threading
from enum import Enum

import common.common_methods as commMethods
from common.checkers.param_checker import LogServiceDownloadChecker
from common.constants.base_constants import LoggerConstants, CommonConstants
from common.constants.product_constants import COLLECT_LOG_SHELL_PATH
from common.file_utils import FileUtils
from common.log.logger import run_log
from common.utils.exception_utils import OperateBaseError
from common.utils.exec_cmd import ExecCmd
from common.utils.system_utils import SystemUtils


class LogsCollectError(OperateBaseError):
    pass


class CollectState(Enum):
    EMPTY = "empty"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"


class LoggerCollect:
    collect_progress = 0
    collect_state = CollectState.EMPTY
    log_lock = threading.Lock()

    def __init__(self):
        self.PercentComplete: int = 0
        self.TaskState: str = ""

    @staticmethod
    def check_external_parameter(checker_class, para_data) -> list:
        try:
            check_ret = checker_class().check(para_data)
        except Exception as err_info:
            message = "Check external parameter failed. Because of Parameter is invalid"
            run_log.error(f"{message} {err_info}")
            return [commMethods.CommonMethods.ERROR, message]

        if not check_ret.success:
            message = "checker_class failed. Because of Parameter is invalid"
            run_log.error(f"{message}")
            return [commMethods.CommonMethods.ERROR, message]

        return [commMethods.CommonMethods.OK, ]

    @staticmethod
    def _get_log_collect_cmd(log_names):
        # 传入参数去重
        log_names = log_names.strip(" ").split(" ")
        return [os.path.join(CommonConstants.OM_WORK_DIR_PATH, "software", "ibma", COLLECT_LOG_SHELL_PATH)] + log_names

    def collect_log(self, request_data):
        # 如果存在，先尝试删除
        FileUtils.delete_file_or_link(LoggerConstants.MONITOR_TMP_COLLECT_LOG)

        if SystemUtils.get_fs_used_status(LoggerConstants.RUN_DIR)[1] < LoggerConstants.RUN_DIR_AVAILABLE_SIZE:
            raise LogsCollectError("the available space is not enough")

        cmd = self._get_log_collect_cmd(request_data.get("name"))
        sign_cmd = ExecCmd.exec_cmd_get_output(cmd, wait=600)
        if sign_cmd[0] != 0:
            raise LogsCollectError("exec log_collect failed")

        # 日志文件超过最大限制，也认为失败
        if os.path.getsize(LoggerConstants.MONITOR_TMP_COLLECT_LOG) > LoggerConstants.COLLECT_LOG_UPPER_LIMIT:
            raise LogsCollectError("Collect log failed because the log size is up to limit.")

    def post_request(self, request_data):
        if self.log_lock.locked():
            run_log.error("log collect is busy")
            return [commMethods.CommonMethods.ERROR, "log collect is busy"]

        with self.log_lock:
            LoggerCollect.collect_progress = 0
            LoggerCollect.collect_state = CollectState.RUNNING
            check_ret = self.check_external_parameter(LogServiceDownloadChecker, request_data)
            if check_ret[0] == commMethods.CommonMethods.ERROR:
                run_log.error("Check input para failed, %s", check_ret)
                return [commMethods.CommonMethods.ERROR, f"Check input para failed, {check_ret}"]
            LoggerCollect.collect_progress = 30

            try:
                self.collect_log(request_data)
            except Exception as err:
                LoggerCollect.collect_state = CollectState.FAILED
                run_log.error("Collect log failed: %s", err)
                FileUtils.delete_file_or_link(LoggerConstants.MONITOR_TMP_COLLECT_LOG)
                return [commMethods.CommonMethods.ERROR, "Collect log failed"]

            run_log.info("Collect log success.")
            LoggerCollect.collect_progress = 100
            LoggerCollect.collect_state = CollectState.SUCCESS
            return [commMethods.CommonMethods.OK, "Collect log success."]

    def get_all_info(self):
        self.PercentComplete = LoggerCollect.collect_progress
        self.TaskState = LoggerCollect.collect_state.value

    def delete_request(self, *args):
        """当redfish拷贝好日志文件后，发送delete请求，删除临时日志文件"""
        with self.log_lock:
            try:
                FileUtils.delete_file_or_link(LoggerConstants.MONITOR_TMP_COLLECT_LOG)
            except Exception as err:
                run_log.error("delete collect_log.tar.gz failed: %s", err)
