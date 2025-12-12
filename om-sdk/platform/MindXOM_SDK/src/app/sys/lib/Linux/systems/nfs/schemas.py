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
import os
import re
from dataclasses import dataclass
from enum import Enum
from functools import partial
from typing import Iterable, Callable, Dict

from common.checkers import LocalIpChecker
from common.file_utils import FileCheck
from common.file_utils import FileCreate
from common.log.logger import run_log
from common.schema import BaseModel, field
from common.utils.app_common_method import AppCommonMethod
from common.utils.singleton import Singleton
from lib.Linux.systems.disk.partition import Partition
from lib.Linux.systems.nfs import errors
from lib.Linux.systems.nfs.cfg_mgr import get_nfs_config_count, NFS_MAX_CFG_NUM, mount_path_already_exists
from lib.Linux.systems.nfs.models import NfsCfg


class Operation(Enum):
    MOUNT = "mount"
    UMOUNT = "umount"


class NfsStatus(Singleton):
    ERROR = "error"
    OK = "ok"
    # 用于异常告警
    STATUS_FILE = "/run/nfs/nfs_status_info"

    def __init__(self):
        # 状态缓存
        self.status_cache: Dict[str, str] = {}

    def set_nfs_status(self, status_key: str, status: str):
        """
        设置nfs的状态，用于告警上报：
            1.挂载成功，写入ok
            2.解挂成功，删除对应记录
            3.监控时：如果状态正常，则设置成ok，否则error
        格式：
            "{cfg.server_ip} {cfg.server_dir} {cfg.local_dir} ={status}"
        """
        self.status_cache[status_key] = status

    def set_status_ok(self, status_key: str):
        self.set_nfs_status(status_key, self.OK)

    def set_status_error(self, status_key: str):
        self.set_nfs_status(status_key, self.ERROR)

    def delete_status(self, status_key: str):
        if status_key in self.status_cache:
            self.status_cache.pop(status_key)

    def save(self):
        if not os.path.exists(os.path.dirname(self.STATUS_FILE)):
            if not FileCreate.create_dir(os.path.dirname(self.STATUS_FILE), 0o700):
                raise errors.MountError("create status path error.")
        if not FileCheck.check_input_path_valid(self.STATUS_FILE):
            raise errors.MountError("status file invalid.")
        with os.fdopen(os.open(self.STATUS_FILE, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as status_file:
            status_file.writelines(f"{key} ={value}" for key, value in self.status_cache.items())


@dataclass
class NfsReq(BaseModel):
    operate: str = field(alias="Type", comment="操作类型")
    server_ip: str = field(alias="ServerIP", comment="NFS服务器IP地址")
    server_dir: str = field(alias="ServerDir", comment="NFS服务器共享路径")
    fs_type: str = field(alias="FileSystem", comment="NFS文件系统")
    mount_path: str = field(alias="MountPath", comment="NFS分区本地挂载路径")

    def __post_init__(self):
        self.validate()
        self.mount_path = os.path.normpath(self.mount_path)

    @staticmethod
    def check_input_path(path):
        if path is None:
            return True
        if len(path) > 256:
            return False
        pattern_name = re.compile(r'[^0-9a-zA-Z_./-]')
        match_name = pattern_name.findall(path)
        if match_name or ".." in path:
            return False
        else:
            return True

    @classmethod
    def from_dict(cls, data: dict):
        if not data:
            raise errors.ParmaError("Parameter is null")
        return super().from_dict(data)

    def validate(self):
        for validator in self.validators():
            validator()

    def validators(self) -> Iterable[Callable]:
        yield self.valid_operate
        yield self.valid_server_ip
        yield self.valid_server_dir
        yield self.valid_fs_type
        if self.operate == Operation.MOUNT.value:
            yield from MountValidator(self.mount_path).validators()
        else:
            yield from UmountValidator(self).validators()

    def valid_operate(self):
        if self.operate not in (item.value for item in Operation):
            raise errors.ParmaError("Unknown nfs operation type")

    def valid_server_ip(self):
        if not self.server_ip or not AppCommonMethod.check_ipv4_format(self.server_ip):
            raise errors.ParmaError("Invalid server ip")

        ret = LocalIpChecker("server_ip").check_dict({"server_ip": self.server_ip})
        if not ret:
            raise errors.ParmaError(f"Invalid server ip: {ret.reason}")

    def valid_server_dir(self):
        if not self.server_dir or not self.check_input_path(self.server_dir):
            raise errors.ParmaError("Server share path invalid: too long or contains invalid char.")

    def valid_fs_type(self):
        if self.fs_type != "nfs4":
            raise errors.FilesystemWrong("Invalid file system type")


class MountValidator:
    def __init__(self, path):
        self.path = path

    @staticmethod
    def not_exceeds_limit():
        if get_nfs_config_count() >= NFS_MAX_CFG_NUM:
            raise errors.CfgExceedsLimit("NFS configuration exceeds limit.")

    def validators(self) -> Iterable[Callable]:
        yield partial(valid_path, self.path)
        yield self.not_exists
        yield partial(is_whitelist_and_permitted, self.path)
        yield self.is_not_subdir
        yield self.not_exceeds_limit

    def not_exists(self):
        if os.path.exists(self.path):
            raise errors.MountPathExisted("Mount path already exists.")
        if mount_path_already_exists(self.path):
            raise errors.MountPathExisted("Mount path already in configs.")

    def is_not_subdir(self):
        # 不能挂载到已挂载目录的子目录中，也不能挂载到已挂载的父目录下。CommonMethods存在父子关系时返回了False
        if not Partition.check_mount_path_is_subdirectory_of_mounted_path(self.path):
            raise errors.SubdirRelation("There is a subdirectory relationship between Mount Path and Mounted path")


class UmountValidator:
    def __init__(self, request: NfsReq):
        self.cfg = NfsCfg(request.server_ip, request.server_dir, request.mount_path, request.fs_type)

    def validators(self) -> Iterable[Callable]:
        yield partial(Partition.check_path_is_permitted, path=self.cfg.local_dir)
        yield self.exists

    def exists(self):
        try:
            existed = mount_path_already_exists(self.cfg.local_dir)
        except Exception as err:
            run_log.error("Query nfs config err. Catch %s", err.__class__.__name__)
            raise errors.UmountPathNotExists("Mount path does not in Nfs config.") from err
        if not existed:
            raise errors.UmountPathNotExists("Mount path does not in Nfs config.")


def valid_path(path):
    if not path:
        raise errors.ParmaError("Mount path is null")
    if not FileCheck.check_input_path_valid(path):
        raise errors.ParmaError("NFS Mount path invalid")


def is_whitelist_and_permitted(path):
    if path == "/var/lib/docker":
        raise errors.MountPathInvalid("Mount path is not in whitelist.")
    if not Partition.check_path_whitelist(path) or not Partition.check_path_is_permitted(path):
        raise errors.MountPathInvalid("Mount path is not in whitelist.")
