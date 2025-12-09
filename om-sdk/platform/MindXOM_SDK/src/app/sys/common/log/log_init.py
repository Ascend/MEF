# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import logging
import os

from common.file_utils import FileCreate
from common.log.log_handlers import WatchedFileSoftLinkHandler


def init_om_logger(log_file_name, log_path, operate_flag=False):

    OMLogger().init_logger(log_name=log_file_name, log_path=log_path, log_type="run")

    if operate_flag:
        OMLogger().init_logger(log_name=log_file_name, log_path=log_path, log_type="operate")


class OMLogger:
    _DATE_FORMAT = "%Y-%m-%d %H:%M:%S"
    _SUPPORT_TYPE_TUPLE = ("run", "operate")

    def init_logger(self, log_name, log_type="run", log_path="", log_level=logging.INFO):
        if not log_path:
            log_path = os.path.join("/var/plog", "om_log")

        log_path = os.path.realpath(log_path)
        if not os.path.exists(log_path):
            if not FileCreate.create_dir(log_path, mode=0o700):
                return False

        if log_type not in self._SUPPORT_TYPE_TUPLE:
            return False

        logger_instance = logging.getLogger(log_type)
        logger_instance.setLevel(log_level)

        run_log_format = f"[%(asctime)s] [%(levelname)s] [{log_name}] " \
                         f"[%(funcName)s] [%(filename)s:%(lineno)d] %(message)s"

        operate_log_format = f"[%(asctime)s] %(message)s"

        log_full_path = os.path.join(log_path, f"{log_name}_{log_type}.log")
        watch_handler = WatchedFileSoftLinkHandler(log_full_path)

        log_file = os.path.realpath(log_full_path)
        FileCreate.create_file(log_file, 0o600)

        if log_type == "run":
            formatter = logging.Formatter(run_log_format, self._DATE_FORMAT)
        else:
            formatter = logging.Formatter(operate_log_format, self._DATE_FORMAT)

        watch_handler.setFormatter(formatter)
        logger_instance.addHandler(watch_handler)
        return True
