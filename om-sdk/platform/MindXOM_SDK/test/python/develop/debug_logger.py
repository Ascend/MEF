import logging
import os

from arguments import Arguments
from common.log.log_constant import LogConstant
from common.log.log_init import init_om_logger


def console_handler(log_type: str) -> logging.StreamHandler:
    fmt = f"[%(asctime)s] [%(levelname)s] [{log_type}] [%(funcName)s] [%(filename)s:%(lineno)d] %(message)s"
    handler = logging.StreamHandler()
    handler.setFormatter(logging.Formatter(fmt, "%Y-%m-%d %H:%M:%S"))
    return handler


def init_logger(log_type: str, args: Arguments):
    """将日志定向到调试用户目录下;并添加定向到console,便于调试时观察日志。"""
    log_path = os.path.join(LogConstant.OM_LOG_DIR, args.user, log_type)
    init_om_logger(log_type, log_path=log_path, operate_flag=True)

    logging.getLogger("run").addHandler(console_handler(f"{log_type}_run"))
    logging.getLogger("operate").addHandler(console_handler(f"{log_type}_operate"))
