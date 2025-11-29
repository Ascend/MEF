# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import inspect

from common.constants.error_code_constants import ErrorCode


class BaseError(Exception):
    CODE: int = 0

    def __init__(self, err_msg: str):
        self.err_msg = err_msg
        # 将触发异常的调用点添加到错误信息中，统一打印错误时，明确触发错误的信息，便于定位。
        info: inspect.FrameInfo = inspect.stack()[1]
        msg = f"{info.function}({info.lineno})-{err_msg}"
        super().__init__(msg)


class UpgradeError(BaseError):
    CODE = ErrorCode.midware_firmware_upgrade_err


class DownloadError(BaseError):
    CODE = ErrorCode.midware_firmware_download_err


class TimeOutError(BaseError):
    CODE = ErrorCode.upgrade_timeout_err_code


class ExternalParmaError(BaseError):
    pass
