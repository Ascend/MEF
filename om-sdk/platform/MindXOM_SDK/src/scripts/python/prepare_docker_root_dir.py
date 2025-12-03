# !/usr/bin/python
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import configparser
import os
import sys
from pathlib import Path

from bin.environ import Env
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.utils.system_utils import SystemUtils


class DockerRootDirPrepare:
    PARTITION = "p8"
    MOUNT_POINT = Path("/var/lib/docker")
    PERSIST_FILE = Path("/home/data/ies/mountCnf_site.ini")

    def execute(self) -> int:
        if not self.valid():
            return 1

        try:
            self.persist()
        except Exception as error:
            run_log.error("persist docker root dir failed, catch %s", error.__class__.__name__)
            return 1

        return 0

    def persist(self):
        if not FileCheck.check_input_path_valid(self.PERSIST_FILE.as_posix()):
            run_log.error("mountCnf_site invalid")
            return

        cfg_parser = configparser.ConfigParser()
        partition = Env().get_partition_by_label(self.PARTITION)
        cfg_parser.add_section(partition.name)
        cfg_parser.set(partition.name, partition.as_posix(), self.MOUNT_POINT.as_posix())

        with os.fdopen(os.open(self.PERSIST_FILE, os.O_WRONLY | os.O_CREAT | os.O_APPEND, 0o640), "a") as file:
            cfg_parser.write(file)
        run_log.info("prepare docker root dir success.")

    def mount_point_or_mount_part_in_persist(self, part: Path) -> bool:
        if not FileCheck.check_path_is_exist_and_valid(self.PERSIST_FILE.as_posix()):
            return False

        parser = configparser.ConfigParser()
        parser.read(self.PERSIST_FILE)
        for section in parser.sections():
            for option in parser.options(section):
                if self.MOUNT_POINT.as_posix() == parser.get(section, option):
                    return True

        # /var/lib/docker不在持久化文件中，需要判断p8是否已被占，占用的情况下返回True
        return parser.has_section(part.name)

    def valid(self) -> bool:
        if not SystemUtils().is_a500:
            run_log.info("not a500, no need prepare.")
            return False

        try:
            partition = Env().get_partition_by_label(self.PARTITION)
        except Exception as err:
            run_log.error("get p8 catch %s", err.__class__.__name__)
            return False

        if self.mount_point_or_mount_part_in_persist(partition):
            run_log.info("mount point already exists , no need prepare.")
            return False

        if not partition.exists() or not partition.is_block_device():
            run_log.warning("p8 not exists.")
            return False

        # 挂载点未挂载
        if self.MOUNT_POINT.is_mount():
            run_log.warning("Docker root dir already mounted.")
            return False

        return True


if __name__ == '__main__':
    sys.exit(DockerRootDirPrepare().execute())
