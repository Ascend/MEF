# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import inspect
import os
from typing import List, Any

from common.constants import error_codes
from common.exception.biz_exception import BizException


class ExceptionUtils:
    @staticmethod
    def exception_process(ex: Exception) -> List[Any]:
        if isinstance(ex, BizException):
            error_code = ex.args[0]
            return [error_code.code, error_code.messageKey]

        return [error_codes.CommonErrorCodes.ERROR_INTERNAL_SERVER.code,
                error_codes.CommonErrorCodes.ERROR_INTERNAL_SERVER.messageKey]


class OperateBaseError(Exception):
    def __init__(self, err_msg: str = ""):
        info: inspect.FrameInfo = inspect.stack()[1]
        self.err_msg = err_msg
        msg = f"Exception position: [{os.path.basename(info.filename)}] [{info.function}:{info.lineno}]  {err_msg}"
        super().__init__(msg)


class OperationCode(object):
    SUCCESS_OPERATION = 0
    FAILED_OPERATION = 1
