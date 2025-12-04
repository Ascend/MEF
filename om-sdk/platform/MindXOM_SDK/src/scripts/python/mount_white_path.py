# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import inspect
import os.path
import signal
import sys
from argparse import ArgumentParser
from enum import Enum
from functools import partial
from pathlib import Path
from typing import Dict, Callable

from common.file_utils import FileCheck
from common.utils.scripts_utils import signal_handler
from lib.Linux.systems.disk import mount_mgr as mgr
from logger import terminal_print

DEFAULT_PATH = ("/opt/mount", "/var/lib/docker")


class WhitelistError(Exception):
    pass


def is_sub_dir(sub, parent):
    if sys.version_info >= (3, 9):
        return Path(sub).is_relative_to(parent)

    try:
        Path(sub).relative_to(parent)
    except ValueError:
        return False
    return True


def add_whitelist_path(path: str):
    # 判断白名单是否已达到最大
    if mgr.get_whitelist_path_count() >= mgr.MAX_WHITE_PATH_CNT:
        raise WhitelistError(
            f"White path count reaches the maximum value({mgr.MAX_WHITE_PATH_CNT}), the whitelist path cannot be added."
        )

    ret = FileCheck.check_input_path_valid(path)
    if not ret:
        raise WhitelistError(f"Input path invalid, {ret.error}")

    # 添加白名单时检查的路径一般是不存在的，需要跳过不存在的目录层级，只检查存在的目录属主是否root
    ret = FileCheck.check_path_is_root(path, skip_no_exists=True)
    if not ret:
        raise WhitelistError(f"Input path invalid, {ret.error}")

    # 校验是否已在白名单
    if mgr.path_in_whitelist(path):
        raise WhitelistError(f"{path} path already exist in whitelist")

    path = os.path.realpath(path)
    err_template = "{} path is a {} path of {} path, the whitelist path cannot contain any relationship."
    for white in mgr.query_whitelist_path():
        # 校验添加的路径是否是已有路径的子路径
        existed_path = os.path.realpath(white.path)
        if is_sub_dir(path, existed_path):
            raise WhitelistError(err_template.format(path, "sub", white.path))

        # 校验添加的路径是否是已有路径的父路径
        if is_sub_dir(existed_path, path):
            raise WhitelistError(err_template.format(path, "parent", white.path))

    mgr.add_whitelist_path(path)
    terminal_print.info("%s added successfully", path)


def delete_whitelist_path(path: str):
    if os.path.normpath(path) in DEFAULT_PATH:
        raise WhitelistError("Cannot delete the default path")

    # 判断路径是不是已挂载
    if Path(path).is_mount():
        raise WhitelistError("Cannot delete a mounted path in whitelist")

    if mgr.delete_mount_white_path(path):
        terminal_print.info("%s deleted successfully", path)
        return

    raise WhitelistError(f"Cannot delete {path}, it does not exist in whitelist")


def display_whitelist_path():
    for white in mgr.query_whitelist_path():
        terminal_print.info(white.path)


def path_in_whitelist(path: str):
    if mgr.path_in_whitelist(path):
        terminal_print.info("%s already exist in whitelist.", path)
        return

    terminal_print.info("%s does not exist in the whitelist.", path)


class Operation(Enum):
    ADD = "add"
    DELETE = "delete"
    CHECK = "check"
    DISPLAY = "display"


def input_args():
    parse = ArgumentParser()
    parse.add_argument("operate", type=str, choices={item.value for item in Operation}, help="操作类型")
    parse.add_argument("-p", "--path", help="路径")
    return parse.parse_args()


OPERATE: Dict[str, Callable] = {
    Operation.ADD.value: add_whitelist_path,
    Operation.DELETE.value: delete_whitelist_path,
    Operation.CHECK.value: path_in_whitelist,
    Operation.DISPLAY.value: display_whitelist_path,
}


def main() -> int:
    args = input_args()
    # args.operate 一定在OPERATE中，否则参数校验就触发异常了
    operate = OPERATE[args.operate]
    if inspect.getfullargspec(operate).args:
        operate = partial(operate, args.path)

    try:
        operate()
    except WhitelistError as error:
        terminal_print.error(error)
        return 1
    except Exception as error:
        terminal_print.error("%s whitelist path catch %s", args.operate, error.__class__.__name__)
        return 1
    return 0


if __name__ == '__main__':
    # 注册退出信号的中断处理函数
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    try:
        sys.exit(main())
    except Exception as err:
        terminal_print.error("catch %s", err.__class__.__name__)
        sys.exit(1)
