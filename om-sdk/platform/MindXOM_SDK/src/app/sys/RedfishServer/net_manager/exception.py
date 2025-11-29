# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import inspect
from pathlib import Path

from common.constants.error_codes import ErrorCode


class NetManagerException(Exception):
    """网管模块异常基类."""

    def __init__(self, err_msg: str, err_code=None):
        self.err_msg = err_msg
        # 用于保存与前端约定的错误码
        self.err_code = err_code
        # 将触发异常的调用点添加到错误信息中，统一打印错误时，明确触发错误的信息，便于定位。
        info: inspect.FrameInfo = inspect.stack()[1]
        msg = f"{info.function}({Path(info.filename).name}:{info.lineno})-{err_msg}"
        super().__init__(msg)


class LockedError(NetManagerException):
    pass


class ValidateParamsError(NetManagerException):
    pass


class DbOperateException(NetManagerException):
    """数据库操作异常类."""
    pass


class DataCheckException(NetManagerException):
    """数据校验异常类."""
    pass


class InvalidCertInfo(DataCheckException):

    def __init__(self, error: ErrorCode):
        super().__init__(err_msg=error.messageKey, err_code=error.code)


class FileCheckException(NetManagerException):
    """文件校验异常类."""
    pass


class InvalidDataException(NetManagerException):
    """无效数据异常类."""
    pass


class FileOperateException(NetManagerException):
    """文件操作异常类."""
    pass


class KmcOperateException(NetManagerException):
    """文件操作异常类."""
    pass


class NetSwitchException(NetManagerException):
    """网管切换操作异常类."""
    pass
