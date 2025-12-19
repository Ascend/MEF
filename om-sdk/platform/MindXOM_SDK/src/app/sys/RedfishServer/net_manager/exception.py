# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import inspect
from pathlib import Path


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
