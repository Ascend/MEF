# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from typing import Any, Tuple
import re
from common.constants.error_codes import ErrorCode


class BizException(RuntimeError):
    error_code = -1
    msg = None
    data = None

    def __init__(self, error_code, *args, **kwargs):
        super(BizException, self).__init__(error_code, *args, **kwargs)


class Exceptions:
    @staticmethod
    def biz_exception(error_code: ErrorCode, *args):
        error_code.messageKey = re.sub(u"\\[.*?]", "[{}]", error_code.messageKey)
        if not Exceptions.check_list_is_null(args):
            error_code.messageKey = error_code.messageKey.format(*args)
        if "[{}]" in error_code.messageKey:
            error_code.messageKey = error_code.messageKey.replace("[{}]", "")
            error_code.messageKey = re.sub(' +', ' ', error_code.messageKey)
        return BizException(error_code, *args)

    @staticmethod
    def check_list_is_null(param_list: Tuple[Any]) -> bool:
        if not param_list:
            return True
        for value in param_list:
            if value:
                return False
        return True
