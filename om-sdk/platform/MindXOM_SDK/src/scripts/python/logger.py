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

DATA_FMT = "%Y-%m-%d %H:%M:%S"


def init_logger(log_type: str, log_file: str, log_format: str) -> logging.Logger:
    logger = logging.getLogger(log_type)
    logger.setLevel(logging.INFO)
    run_logger = logging.FileHandler(log_file)
    run_logger.setFormatter(logging.Formatter(log_format, DATA_FMT))
    logger.addHandler(run_logger)
    return logger


def init_terminal_logger(log_type: str, log_format: str) -> logging.Logger:
    """将日志信息输出到终端"""
    logger = logging.getLogger(log_type)
    logger.setLevel(logging.INFO)
    terminal_logger = logging.StreamHandler()
    terminal_logger.setFormatter(logging.Formatter(log_format, DATA_FMT))
    logger.addHandler(terminal_logger)
    return logger


def get_upgrade_run_logger() -> logging.Logger:
    run_log_format = "[%(asctime)s] [%(levelname)s] [%(funcName)s] [%(filename)s:%(lineno)d] %(message)s"
    log_file = '/var/plog/upgrade.log'
    return init_logger("log_run", log_file, run_log_format)


def get_terminal_print_logger() -> logging.Logger:
    terminal_output_format = "%(message)s"
    return init_terminal_logger("terminal", terminal_output_format)


upgrade_log = get_upgrade_run_logger()
terminal_print = get_terminal_print_logger()
