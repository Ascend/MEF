# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import time
from datetime import datetime
from typing import AnyStr

from common.constants.base_constants import CommonConstants


class DateUtils:
    @staticmethod
    def get_format_time(time_float: float) -> AnyStr:
        """
        获取当前格式化的时间
        @return: 时间字符串
        """
        return time.strftime(CommonConstants.STR_DATE_FORMAT, time.localtime(time_float))

    @staticmethod
    def get_time(str_datetime: AnyStr) -> datetime:
        """
        获取当前格式化的时间
        @return: 时间字符串
        """
        return datetime.strptime(str_datetime, CommonConstants.STR_DATE_FORMAT)

    @staticmethod
    def default_time():
        """
        获取当前格式化的时间，作为ORM中的default函数用，需要无参
        @return: 时间字符串
        """
        return DateUtils.get_format_time(time.time())
