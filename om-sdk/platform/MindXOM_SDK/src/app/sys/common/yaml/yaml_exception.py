# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import inspect
from pathlib import Path


class YamlException(Exception):
    """YAML模块异常类."""

    def __init__(self, err_msg: str):
        self.err_msg = err_msg
        # 将触发异常的调用点添加到错误信息中，统一打印错误时，明确触发错误的信息，便于定位。
        info: inspect.FrameInfo = inspect.stack()[1]
        msg = f"{info.function}({Path(info.filename).name}:{info.lineno})-{err_msg}"
        super().__init__(msg)
