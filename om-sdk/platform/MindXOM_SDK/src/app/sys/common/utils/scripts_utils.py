# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import sys
from typing import NoReturn


def signal_handler(sig_num=None, frame=None) -> NoReturn:
    """
    信号处理函数
    :param sig_num: 信号值
    :param frame: 栈帧
    :return: None
    """
    sys.exit(1)
