# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import logging
import os
import pwd

from common.log.log_constant import LogConstant
from common.log.log_init import init_om_logger
from common.utils.singleton import Singleton


class ProcessLogger(Singleton):
    settings = {
        LogConstant.NGINX_USER: {
            "log_file_name": LogConstant.NGINX_MODULE_NAME,
            "log_path": os.path.join(LogConstant.OM_LOG_DIR, LogConstant.NGINX_MODULE_NAME),
            "operate_flag": False,
        },
        LogConstant.REDFISH_USER: {
            "log_file_name": LogConstant.REDFISH_MODULE_NAME,
            "log_path": os.path.join(LogConstant.OM_LOG_DIR, LogConstant.REDFISH_MODULE_NAME),
            "operate_flag": True,
        },
        LogConstant.MANAGER_USER: {
            "log_file_name": LogConstant.MANAGER_MODULE_NAME,
            "log_path": os.path.join(LogConstant.OM_LOG_DIR, LogConstant.MANAGER_MODULE_NAME),
            "operate_flag": True,
        }
    }

    def __init__(self):
        process_name = pwd.getpwuid(os.getuid()).pw_name
        if process_name in self.settings:
            init_om_logger(**self.settings[process_name])

    @staticmethod
    def get_custom_logger(name):
        return logging.getLogger(name) if name in logging.Logger.manager.loggerDict else None


def init_terminal_logger(log_type: str, log_format: str) -> logging.Logger:
    """将日志信息输出到终端"""
    logger = logging.getLogger(log_type)
    logger.setLevel(logging.INFO)
    terminal_logger = logging.StreamHandler()
    terminal_logger.setFormatter(logging.Formatter(log_format, "%Y-%m-%d %H:%M:%S"))
    logger.addHandler(terminal_logger)
    return logger


def get_terminal_print_logger() -> logging.Logger:
    terminal_output_format = "%(message)s"
    return init_terminal_logger("shell", terminal_output_format)


# 不同进程中Logger的单例，operate_log由operate_flag决定是否初始化，可能不存在，根据进程初始化情况酌情使用
process_log_object = ProcessLogger()
run_log = process_log_object.get_custom_logger("run")
operate_log = process_log_object.get_custom_logger("operate")

# 后台脚本输出到终端的logger
terminal_print = get_terminal_print_logger()
