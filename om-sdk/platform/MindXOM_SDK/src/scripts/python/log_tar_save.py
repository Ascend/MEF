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
import argparse
import os
import sys
import tarfile
import time
from collections import namedtuple
from pathlib import Path
from typing import NoReturn
from typing import Type

from common.file_utils import FileCheck
from common.file_utils import FileCreate
from common.file_utils import FilePermission as Chmod
from common.file_utils import FileUtils
from common.log.logger import run_log
from common.utils.exception_utils import OperateBaseError

LogParam = namedtuple("LogParam", ["src_dir", "log_base_name", "dest_dir", "max_num"])
MAX_LOOP_NUM = 1000


class DumpLogError(OperateBaseError):
    pass


def parse_input_param() -> Type[LogParam]:
    """解析校验参数"""
    parse = argparse.ArgumentParser()
    parse.add_argument("src_dir", type=check_path, help="源日志目录路径")
    parse.add_argument("log_base_name", type=str, help="源日志文件名")
    parse.add_argument("dest_dir", type=check_path, help="目标日志目录路径")
    parse.add_argument("max_num", type=int, help="目标目录允许的最大日志文件数量")
    return parse.parse_args(namespace=LogParam)


def check_path(path: str) -> str:
    ret = FileCheck.check_input_path_valid(path)
    if not ret:
        raise DumpLogError(f"The log file is not exists, because {ret.error}")
    return path


def del_old_log_storage(dest_dir: str, max_num: int, log_base_name: str) -> NoReturn:
    """转储日志文件个数如果超过最大值，删除修改日期最久的文件"""
    # 如果目录不存在，则不需要删除多余文件，直接创建目录即可
    if not os.path.exists(dest_dir):
        run_log.info("The log dump dir is not exist, creat it.")
        ret = FileCreate.create_dir(dest_dir, 0o750)
        if not ret:
            raise DumpLogError(f"Create dir {dest_dir} failed, because {ret.error}")
        return

    # glob("*") 返回当前目录下的【所有文件夹】和【所有文件】
    log_list = []
    for file in Path(dest_dir).glob(f"{log_base_name}*"):
        if len(log_list) > MAX_LOOP_NUM:
            break

        if file.is_file():
            log_list.append(file)

    # 按照修改日期排序，排在最后面的是修改日期最久的文件
    log_list.sort(key=lambda x: os.path.getmtime(x), reverse=True)
    while log_list and len(log_list) >= max_num:
        # 删除修改日期最久的文件
        log_file = log_list.pop()
        try:
            FileUtils.delete_file_or_link(log_file)
        except Exception as err:
            raise DumpLogError(f"delete {log_file.name} failed, because {err}") from err


def tar_save_log(dest_dir: str, src_dir: str, log_base_name: str) -> NoReturn:
    """将日志转储打包压缩"""
    now_time = time.strftime("%Y%m%d_%H%M%S")
    dest_tar_path = os.path.join(dest_dir, f"{log_base_name}_{now_time}.tar.gz")
    try:
        with tarfile.open(dest_tar_path, "w:gz") as tar:
            # logrotate 默认情况下将日志文件xxxx.log转储为 xxxx.log.1，因此需要添加 xxxx.log.1
            tar.add(os.path.join(src_dir, f"{log_base_name}.log.1"))
    except Exception as err:
        raise DumpLogError(f"Failed to package log file, because unknown error {err}") from err

    ret = Chmod.set_path_permission(dest_tar_path, 0o440)
    if not ret:
        raise DumpLogError(f"chmod file {dest_tar_path} failed, because {ret.error}")


def main() -> NoReturn:
    args = parse_input_param()
    run_log.info("Start dump log %s...", args.log_base_name)
    del_old_log_storage(args.dest_dir, args.max_num, args.log_base_name)
    tar_save_log(args.dest_dir, args.src_dir, args.log_base_name)
    run_log.info("Dump log success.")


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        run_log.error("log dump failed, because %s. Error Type %s", error, error.__class__.__name__)
        sys.exit(1)
    sys.exit(0)
