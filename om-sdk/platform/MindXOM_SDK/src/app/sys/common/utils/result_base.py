# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
"""
功 能：
版权信息：华为技术有限公司，版本所有(C) 2021-2029
"""


class Result:
    def __init__(self, result: bool, data=None, err_msg: str = "", err_code: str = ""):
        self._result = result
        self._data = data
        self._err_msg = err_msg
        self._err_code = err_code

    def __bool__(self):
        return self._result

    def __str__(self):
        return f"result: {self._result}, msg: {self._err_msg}."

    @property
    def data(self):
        return self._data

    @property
    def error(self) -> str:
        return self._err_msg

    @property
    def error_code(self):
        return self._err_code
