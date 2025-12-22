# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.


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
