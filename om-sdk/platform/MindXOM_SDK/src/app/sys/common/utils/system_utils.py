# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os
import shutil
from functools import cached_property
from pathlib import Path
from typing import Union, Tuple

from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.utils.exec_cmd import ExecCmd
from common.utils.singleton import Singleton


class SystemUtils(Singleton):
    KB_TO_MB = 1024

    @cached_property
    def is_a500(self) -> bool:
        return self.get_model_by_npu_smi() in CommonConstants.A500_MODELS

    @staticmethod
    def get_available_size(file_systems_path: str) -> int:
        """获取可使用空间大小"""

        vfs = os.statvfs(file_systems_path)
        return vfs.f_bsize * vfs.f_bavail

    @staticmethod
    def get_fs_used_status(fs_path: Union[str, Path]) -> Tuple[int, int]:
        """
        计算指定文件系统的使用情况，返回单位Bytes

             - 总容量(KB) = f_bsize * f_blocks / 1024 【1k - blocks】
             - 使用量(KB) = f_bsize * (f_blocks - f_bfree) / 1024 【Used】
             - 有效容量(KB) = f_bsize * f_bavail / 1024 【Available】
             【】的内容，对应df命令的相同名称的列
        """
        vfs = os.statvfs(fs_path)
        used = vfs.f_bsize * (vfs.f_blocks - vfs.f_bfree)
        avail = vfs.f_bsize * vfs.f_bavail
        return used, avail

    @staticmethod
    def get_fs_used_size(fs_path: Union[str, Path]) -> float:
        """
        计算指定路径的使用大小，通过命令du获取, du命令默认返回kb，所以除以1024
        """
        cmd = f"du -s {fs_path}"
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd)
        if ret[0] != 0 or not ret[1]:
            raise ValueError(f"Get used size failed by {cmd}.")
        return float(ret[1].split()[0]) / SystemUtils.KB_TO_MB

    @staticmethod
    def get_valid_command(name: str) -> str:
        cmd = shutil.which(name)
        if not FileCheck.check_path_is_root(cmd):
            return ""
        return cmd

    @staticmethod
    def get_model_by_npu_smi() -> str:
        """通过npu-smi info -t product -i 0从环境中获取板型号"""
        cmd_str = "npu-smi info -t product -i 0 | grep 'Product Type'"
        status, data = ExecCmd.exec_cmd_use_pipe_symbol(cmd_str)
        if status != 0 or not data:
            return ""

        return data.split(":")[-1].strip()

    @staticmethod
    def is_tmpfs_type(path):
        """判断当前路径是否为内存文件系统"""
        cmd = f"df {path} | grep tmpfs"
        status, data = ExecCmd.exec_cmd_use_pipe_symbol(cmd)
        return not (status != 0) and not (status == 0 and not data)

    @staticmethod
    def get_sn_by_npu_smi() -> str:
        """通过npu-smi info获取序列号"""
        npu_smi_cmd = "npu-smi info -t board -i 0 | grep 'Serial Number'"
        status, data = ExecCmd.exec_cmd_use_pipe_symbol(npu_smi_cmd, wait=3)
        if status != 0 or not data:
            raise ValueError(f"Get serial number by {npu_smi_cmd} failed.")

        return data.split(":")[-1].strip()

    @staticmethod
    def get_npu_version_by_npu_smi() -> str:
        """通过npu-smi info获取npu版本号"""
        npu_smi_cmd = "npu-smi info -t board -i 0 | grep 'Software Version'"
        status, data = ExecCmd.exec_cmd_use_pipe_symbol(npu_smi_cmd, wait=3)
        if status != 0 or not data:
            raise ValueError(f"Get npu version by {npu_smi_cmd} failed.")

        return data.split(":")[-1].strip()

    @staticmethod
    def get_power_by_npu_smi() -> str:
        """通过npu-smi info获取npu power"""
        npu_smi_cmd = "npu-smi info -t power -i 0 | grep Power"
        status, data = ExecCmd.exec_cmd_use_pipe_symbol(npu_smi_cmd, wait=3)
        if status != 0 or not data:
            raise ValueError(f"Get npu power by {npu_smi_cmd} failed.")

        return data.split(":")[-1].strip()

    @staticmethod
    def get_mcu_version_by_npu_smi() -> str:
        """通过npu-smi info获取mcu版本"""
        npu_smi_cmd = "npu-smi upgrade -b mcu -i 0 | grep 'Version'"
        status, data = ExecCmd.exec_cmd_use_pipe_symbol(npu_smi_cmd, wait=3)
        if status != 0 or not data:
            raise ValueError(f"Get serial number by {npu_smi_cmd} failed.")

        return data.split(":")[-1].strip()

    @staticmethod
    def get_board_id_by_npu_smi() -> str:
        """通过npu-smi info获取board id"""
        npu_smi_cmd = "npu-smi info -t board -i 0 | grep 'Board ID'"
        status, data = ExecCmd.exec_cmd_use_pipe_symbol(npu_smi_cmd, wait=3)
        if status != 0 or not data:
            raise ValueError(f"Get serial number by {npu_smi_cmd} failed.")

        return f"0x{int(data.split(':')[-1].strip(), 16):04x}"

    @staticmethod
    def get_npu_calculate_ability_by_npu_smi() -> str:
        """通过npu-smi info获取board id"""
        npu_smi_cmd = "npu-smi info -t nve-level -i 0 -c 0 | grep 'nve level'"
        status, data = ExecCmd.exec_cmd_use_pipe_symbol(npu_smi_cmd, wait=3)
        if status != 0 or not data:
            raise ValueError(f"Get serial number by {npu_smi_cmd} failed.")

        num_str = ""
        for char in data.split(':')[-1].strip():
            if not char.isdigit():
                break
            num_str += char

        return f"{num_str}TOPS" if num_str else "8TOPS"
