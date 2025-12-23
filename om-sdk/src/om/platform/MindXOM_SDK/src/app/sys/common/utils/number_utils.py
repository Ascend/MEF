# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import sys
from typing import AnyStr

from common.constants import error_codes
from common.exception.biz_exception import Exceptions


class NumberUtils(object):

    MAX_LEN = 256

    @staticmethod
    def is_int_maxsize(param, prompt: AnyStr = None) -> int:
        """
        判断给定的参数是否超过最大的整数范围
        :param param: 要判断的参数
        :param prompt: 提示的内容信息
        :return: 处理后的整数
        """
        # 防止0的问题
        param = NumberUtils._is_int(param, prompt)
        if param > sys.maxsize:
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_PARAM_RANGE_WRONG, prompt)
        return param

    @staticmethod
    def is_positive_int(param, prompt: AnyStr = None) -> int:
        """
        判断给定的参数是否超过最大的整数范围
        :param param: 要判断的参数
        :param prompt: 提示的内容信息
        :return: 处理后的整数
        """
        # 防止0的问题
        param = NumberUtils.is_int_maxsize(param, prompt)
        if param < 0:
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_TYPE, prompt)
        return param

    @staticmethod
    def in_range(target_param, start_param: int, end_param: int, prompt):
        """
        判断指定的参数是否在给定的范围中。
        :param prompt:
        :param target_param: 要比较的参数
        :param start_param: 开始的范围
        :param end_param: 结束的范围
        :return:
        """
        target_param = NumberUtils._is_int(target_param, prompt)
        start_param = NumberUtils._is_int(start_param, prompt)
        end_param = NumberUtils._is_int(end_param, prompt) + 1
        if start_param <= target_param < end_param:
            return True
        return False

    @staticmethod
    def _is_int(param, prompt: AnyStr = None):
        """
        判断输入参数是否为整数类型
        :param param: 要判断的参数
        :param prompt: 提示信息
        :return: 获取的整数信息
        """
        # 防止0的问题
        if isinstance(param, int):
            return param
        if not param:
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST, prompt)
        param = str(param)
        if not param.isdigit():
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_TYPE, prompt)
        if len(param) > NumberUtils.MAX_LEN:
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_PARAM_RANGE_WRONG, prompt)
        return int(param)
