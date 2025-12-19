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
import os.path
import shlex
import shutil
from pathlib import Path
from typing import List

from sqlalchemy import Column, String, Integer

from common.db.base_models import Base
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd


class NfsCfg(Base):
    __tablename__ = "nfs_cfg"

    id = Column(Integer, primary_key=True)
    server_ip = Column(String, comment="NFS服务器IP地址")
    server_dir = Column(String, comment="NFS服务器共享路径")
    local_dir = Column(String, comment="NFS分区本地挂载路径")
    fs_type = Column(String, comment="NFS文件系统")

    def __init__(self, server_ip, server_dir, local_dir, fs_type):
        self.server_ip = server_ip
        self.server_dir = os.path.normpath(server_dir)
        self.local_dir = os.path.normpath(local_dir)
        self.fs_type = fs_type

    @property
    def files_system(self):
        return f"{self.server_ip}:{self.server_dir}"

    @property
    def status_key_string(self):
        return f"{self.server_ip} {self.server_dir} {self.local_dir}"

    def to_dict(self):
        response = {
            "ServerIP": self.server_ip,
            "ServerDir": self.server_dir,
            "MountPath": self.local_dir,
            "FileSystem": self.fs_type,
            "CapacityBytes": "NA",
            "FreeBytes": "NA",
            "Status": "error",
        }
        if nfs_server_valid(self.local_dir) and get_nfs_status(self):
            usage = shutil.disk_usage(self.local_dir)
            response.update(**{"CapacityBytes": usage.total, "FreeBytes": usage.free, "Status": "ok"})
        return response


def nfs_server_valid(local_dir: str, wait=10):
    # df + 超时时间判断服务端是否正常
    ret, msg = ExecCmd.exec_cmd_get_output((cmd_constants.OS_CMD_DF, "-T", shlex.quote(local_dir)), wait=wait)
    if ret != 0:
        run_log.error("nfs server may be error. %s", msg)
        return False
    return True


def get_nfs_status(cfg: NfsCfg) -> bool:
    """
    判断NFS挂载状态
    """
    if not os.path.exists(cfg.local_dir) or not Path(cfg.local_dir).is_mount():
        return False

    # 判断挂载信息是否一致
    ret, msg = ExecCmd.exec_cmd_use_pipe_symbol(
        f"{cmd_constants.OS_CMD_DF} -T {shlex.quote(cfg.local_dir)} | "
        f"{cmd_constants.OS_CMD_GREP} {shlex.quote(cfg.files_system)}"
    )
    if ret != 0:
        run_log.error("Get nfs mounted info failed, %s", msg)
        return False

    mount_info: List[str] = msg.split(" ")
    if len(mount_info) < 3:
        run_log.error("Nfs mounted info invalid.")
        return False

    return all(loc.strip().startswith(his) for loc, his in zip(mount_info[:2], (cfg.files_system, cfg.fs_type)))
