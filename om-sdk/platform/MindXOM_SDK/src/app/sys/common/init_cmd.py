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

from common.file_utils import FileCheck
from common.log.logger import run_log
from common.utils.singleton import Singleton


class DefaultCmdConstants(Singleton):
    """为兼容多种操作系统，动态获取相关命令信息，默认为EulerOS2.0。"""

    OS_NAME = "EulerOS"
    OS_VERSION_ID = "2.0"
    OS_CMD_ARPING = "/usr/bin/arping"
    OS_CMD_BLKID = "/usr/sbin/blkid"
    OS_CMD_CAT = "/usr/bin/cat"
    OS_CMD_CHATTR = "/usr/bin/chattr"
    OS_CMD_CHMOD = "/usr/bin/chmod"
    OS_CMD_CHOWN = "/usr/bin/chown"
    OS_CMD_CP = "/usr/bin/cp"
    OS_CMD_CRACKLIB_UNPACKER = "/usr/sbin/cracklib-unpacker"
    OS_CMD_CUT = "/usr/bin/cut"
    OS_CMD_DF = "/usr/bin/df"
    OS_CMD_ECHO = "/usr/bin/echo"
    OS_CMD_ETHTOOL = "/usr/sbin/ethtool"
    OS_CMD_FDISK = "/usr/sbin/fdisk"
    OS_CMD_GREP = "/usr/bin/grep"
    OS_CMD_HOSTNAME = "/usr/bin/hostname"
    OS_CMD_IFCONFIG = "/usr/sbin/ifconfig"
    OS_CMD_IP = "/usr/sbin/ip"
    OS_CMD_LS = "/usr/bin/ls"
    OS_CMD_MIIPHY = "/usr/local/bin/miiphy"
    OS_CMD_MOUNT = "/usr/bin/mount"
    OS_CMD_MV = "/usr/bin/mv"
    OS_CMD_NETSTAT = "/usr/bin/netstat"
    OS_CMD_NTPDATE = "/usr/sbin/ntpdate"
    OS_CMD_PARTED = "/usr/sbin/parted"
    OS_CMD_PGREP = "/usr/bin/pgrep"
    OS_CMD_PING = "/usr/bin/ping"
    OS_CMD_PS = "/usr/bin/ps"
    OS_CMD_RM = "/usr/bin/rm"
    OS_CMD_ROUTE = "/usr/sbin/route"
    OS_CMD_SED = "/usr/bin/sed"
    OS_CMD_SETSID = "/usr/bin/setsid"
    OS_CMD_SYSTEMCTL = "/usr/bin/systemctl"
    OS_CMD_SQLITE3 = "/usr/bin/sqlite3"
    OS_CMD_TIMEDATECTL = "/usr/bin/timedatectl"
    OS_CMD_UMOUNT = "/usr/bin/umount"
    OS_CMD_AWK = "/usr/bin/gawk"
    OS_CMD_STAT = "/usr/bin/stat"
    OS_CMD_REALPATH = "/usr/bin/realpath"
    OS_CMD_READLINK = "/usr/bin/readlink"
    OS_CMD_ID = "/usr/bin/id"
    OS_CMD_WHO = "/usr/bin/who"
    OS_CMD_NPU_SMI = "/usr/local/sbin/npu-smi"
    OS_CMD_PARTPROBE = "/usr/sbin/partprobe"
    OS_CMD_MKE2FS = "/usr/sbin/mke2fs"
    OS_CMD_LOGROTATE = "/usr/sbin/logrotate"
    OS_CMD_CRACKLIB_FORMAT = "/usr/sbin/cracklib-format"
    OS_CMD_CRACKLIB_PACKER = "/usr/sbin/cracklib-packer"
    OS_CMD_SMARTCTL = "/usr/sbin/smartctl"
    OS_CMD_IES_TOOL = "/usr/local/bin/ies_tool"
    OS_CMD_KILL = "/usr/bin/kill"
    OS_CMD_MKDIR = "/usr/bin/mkdir"
    OS_CMD_BASH = "/usr/bin/bash"
    OS_CMD_MII_TOOL = "/usr/sbin/mii-tool"
    OS_CMD_HOSTNAMECTL = "/usr/bin/hostnamectl"
    OS_CMD_USERDEL = "/usr/sbin/userdel"
    OS_CMD_GROUPDEL = "/usr/sbin/groupdel"
    OS_CMD_BLOCKDEV = "/usr/local/bin/blockdev"

    def __init__(self):
        self.init_cmd_config()

    @staticmethod
    def init_cmd_config():
        cmd_dict = {}
        cmd_config_file = "/home/data/config/os_cmd.conf"

        # 校验配置文件
        ret = FileCheck.check_path_is_exist_and_valid(cmd_config_file)
        if not ret:
            err_msg = f"{cmd_config_file} path invalid : {ret.error}"
            run_log.error("Cmd config initialize failed, %s", err_msg)
            return [1, "cmd config initialize failed."]

        # 读取配置文件中命令信息
        try:
            with open(cmd_config_file) as file:
                for line in file:
                    name_cmd = line.strip().split("=")
                    if len(name_cmd) != 2:
                        continue
                    cmd_dict[name_cmd[0]] = name_cmd[1]
        except Exception as err:
            run_log.error("Read cmd config failed. %s", err)
            return [1, "read cmd config failed."]

        # 校验命令是否配置
        for name, cmd in cmd_dict.items():
            if name in ["OS_NAME", "OS_VERSION_ID"]:
                setattr(DefaultCmdConstants, name, cmd)
                continue

            if not cmd or not os.path.exists(cmd):
                run_log.warning("Cmd not exist: [%s-%s], please check", name, cmd)
                continue

            if not FileCheck.check_is_link(cmd):
                run_log.warning("Cmd is link file: [%s-%s], please check", name, cmd)
                continue

            except_names = (
                "OS_CMD_CRACKLIB_UNPACKER", "OS_CMD_AWK", "OS_CMD_NPU_SMI", "OS_CMD_CRACKLIB_FORMAT",
                "OS_CMD_CRACKLIB_PACKER", "OS_CMD_MII_TOOL",
            )
            cmd_str = cmd.split("/")[-1]
            name_str = name.split("_")[-1]
            if name_str.lower() != cmd_str and name not in except_names:
                run_log.warning("Cmd not same: [%s-%s], please check", name, cmd)
                continue

            if not FileCheck.check_path_mode_owner_group(cmd, user="root", group="root"):
                run_log.warning("Cmd owner invalid: [%s-%s], please check", name, cmd)
                continue

            setattr(DefaultCmdConstants, name, cmd)

        run_log.info("Init cmd config success.")
        return [0, "init cmd config success."]


cmd_constants = DefaultCmdConstants()
