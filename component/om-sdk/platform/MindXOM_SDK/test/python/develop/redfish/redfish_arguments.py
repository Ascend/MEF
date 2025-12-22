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
import os
import socket
from argparse import ArgumentParser

from arguments import Arguments, add_common_arguments


class RedfishArguments(Arguments):
    # .so目录，需要配置到 Run/Debug Configurations > run_main > Environment variables中。否则使用kmc加载.so文件失败
    LD_LIBRARY_PATH = "/usr/local/mindx/MindXOM/software/RedfishServer/lib/c:/usr/local/mindx/MindXOM/lib"

    https: bool
    auth: bool
    port: int

    @property
    def ibma_edge_db_path(self):
        basename, ext = os.path.splitext("/tmp/redfish_edge.db")
        return "".join(("_".join((basename, self.user)), ext))

    @property
    def host(self) -> str:
        with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as sock:
            sock.connect(("8.8.8.8", 80))
            return sock.getsockname()[0]


def add_redfish_arguments(parse: ArgumentParser):
    """redfish启动独有配置入参"""
    parse.add_argument("--port", "-p", type=int, required=True, help="Flask启动端口，同一个调试环境需要唯一，避免冲突")
    parse.add_argument("--auth", "-a", type=bool, required=False, default=False, help="是否启用鉴权")
    parse.add_argument("--https", "-s", type=bool, required=False, default=False, help="是否启用https")


def args_parse() -> RedfishArguments:
    parse = ArgumentParser(description="开发调试入参")
    add_common_arguments(parse)
    add_redfish_arguments(parse)
    return parse.parse_args(namespace=RedfishArguments())
