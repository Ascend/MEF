# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from typing import NoReturn

from fd_msg_process.midware_urls import MidwareUris


def add_midware_route() -> NoReturn:
    """
        功能描述：注册云边协同topic的处理函数的路由
        extend_mid_ware_add_route: 扩展的注册处理函数路由的函数路径
    """
    MidwareUris.mid_ware_add_route()
