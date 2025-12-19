# !/usr/bin/python
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
import configparser
import itertools
import os
import pwd
import shlex
import shutil
import threading
from functools import partial
from pathlib import Path
from typing import List, NoReturn, Union, Iterable

from bin.environ import Env
from common.checkers.param_checker import CreatePatitionChecker
from common.common_methods import CommonMethods
from common.constants.error_codes import PartitionErrorCodes
from common.file_utils import FileCheck
from common.file_utils import FilePermission
from common.file_utils import FileUtils
from common.init_cmd import cmd_constants
from common.log.logger import operate_log
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.common_check import CommonCheck
from common.utils.exec_cmd import ExecCmd
from common.utils.system_utils import SystemUtils
from lib.Linux.mef.mef_info import MefInfo
from lib.Linux.systems.disk import contants as const
from lib.Linux.systems.disk.device_loader import DeviceLoader
from lib.Linux.systems.disk.errors import (DevError, CreatePartFailed, DelPartFailed, PartMounted, PathNotInWhite,
                                           MountPathInvalid, MountFailed, PartNotMount, UmountDockerFailed,
                                           UmountFailed, ConvertGptError, MountPathExisted)
from lib.Linux.systems.disk.mount_mgr import add_whitelist_paths, path_in_whitelist
from lib.Linux.systems.disk.schema import Disk, Part, SubSystem, DevAction
from lib.Linux.systems.disk.storage import Storage
from lib.Linux.systems.disk.udev_monitor import (udev_monitor, terminate_by_gpt, terminate_while_add_part,
                                                 terminate_by_path)
from lib.Linux.systems.nfs.cfg_mgr import query_nfs_configs

ERROR_MESSAGE = {
    "param_wrong": [PartitionErrorCodes.ERROR_PARAMETER_INVALID.code,
                    PartitionErrorCodes.ERROR_PARAMETER_INVALID.messageKey],
    "partition_name_wrong": [PartitionErrorCodes.ERROR_PARTITION_NAME.code,
                             PartitionErrorCodes.ERROR_PARTITION_NAME.messageKey],
    "busy": [PartitionErrorCodes.OPERATE_IS_BUSY.code, PartitionErrorCodes.OPERATE_IS_BUSY.messageKey],
    "operator_illegal": [PartitionErrorCodes.ERROR_ROOT_OPERATE_ILLEGAL.code,
                         PartitionErrorCodes.ERROR_ROOT_OPERATE_ILLEGAL.messageKey],
    "not_enough_space": [
        PartitionErrorCodes.ERROR_SPACE_NOT_ENOUGH.code,
        PartitionErrorCodes.ERROR_SPACE_NOT_ENOUGH.messageKey,
    ],
    "disk_not_exist": [
        PartitionErrorCodes.ERROR_DRIVE_LETTER_NOT_EXISTED.code,
        PartitionErrorCodes.ERROR_DRIVE_LETTER_NOT_EXISTED.messageKey,
    ],
    "partition_num_error": [
        PartitionErrorCodes.ERROR_PARTITION_NUM.code,
        PartitionErrorCodes.ERROR_PARTITION_NUM.messageKey,
    ],
}


class Partition:
    PERSIST_CFG: str = "/home/data/ies/mountCnf_site.ini"
    PT_TYPE = "gpt"
    FS_TYPE = "ext4"
    WAIT = 30  # 执行分区或磁盘加载的最长等待秒数，根据u-disk实际经验数据设置
    exec_cmd = partial(ExecCmd.exec_cmd_get_output)
    lock = threading.Lock()

    def __init__(self):
        """
        功能描述：初始化函数
        参数：
        返回值：无
        异常描述：NA
        """
        self.Name = None
        self.CapacityBytes = None
        self.Device = None
        self.FreeBytes = None
        self.Primary = False
        self.FileSystem = None
        self.MountPath = None
        self.Status = None
        self.DeviceName = None
        self.Location = None
        self.items = []
        # 缓存容器，采用列表用作保序
        self.part_cache: List[str] = []
        self.disk_cache: List[str] = []
        self._refresh_cache()

    @staticmethod
    def write_config_file(config_file: str, cfg_parser: configparser.ConfigParser) -> NoReturn:
        if not FileCheck.check_is_link(config_file):
            run_log.error('Check %s is link file.', config_file)
            return

        try:
            with os.fdopen(os.open(config_file, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o640), "w+") as file:
                cfg_parser.write(file)
        except Exception as err:
            run_log.error(err)

        # 父目录权限为700, os.fdopen指定权限也无法把文件所属组权限设置成功, 此处单独再设置一次权限
        FilePermission.set_path_permission(config_file, 0o640)

    @staticmethod
    def remove_mount_path(mount_path: str) -> NoReturn:
        if not os.path.exists(mount_path):
            return

        try:
            FileUtils.delete_full_dir(mount_path)
        except Exception as err:
            run_log.warning("remove mount path failed : %s", err)

    @staticmethod
    def init_whitelist() -> NoReturn:
        result = ['/opt/mount', '/var/lib/docker']
        try:
            result.extend((cfg.local_dir for cfg in query_nfs_configs()))
            b_ok = Partition.add_whitelist(result)
        except Exception as ex:
            run_log.error("init white list exception!mount_path: %s", ex)
            return

        if b_ok:
            run_log.info("init white list success!")
        else:
            run_log.error("init white list failed!mount_path: %s", result)

    @staticmethod
    def add_whitelist(mount_path_list: List[str]) -> bool:
        euid = os.geteuid()
        if euid != pwd.getpwnam('root').pw_uid:
            run_log.error("The process permission is incorrect.")
            return False

        if not mount_path_list:
            run_log.error("Mount path list is empty.")
            return False

        for path in mount_path_list:
            ret = FileCheck.check_input_path_valid(path)
            if not ret:
                run_log.error("Path check failed, {}".format(ret.error))
                return False

        try:
            add_whitelist_paths(*set(mount_path_list))
        except Exception as err:
            run_log.error("Add white paths failed, catch %s", err.__class__.__name__)
            return False

        return True

    @staticmethod
    def check_path_whitelist(path):
        """
        校验指定的路径是否在白名单里
        :param path:
        :return:
        """

        ret = FileCheck.check_input_path_valid(path)
        if not ret:
            run_log.warning("path check failed, %s", ret.error)
            return False

        try:
            return path_in_whitelist(path)
        except Exception as err:
            run_log.error("query mount white path error, catch %s", err.__class__.__name__)
            return False

    @staticmethod
    def check_path_is_permitted(path) -> bool:
        """
        校验指定的路径是否是允许挂载的目录
        @param path: 挂在路径
        @return: True 是允许挂载的路径，否则，False
        """
        white_paths = ("/opt/", "/home/")
        black_paths = (
            "/home/data/cert_backup/",
            "/home/data/",
            "/home/log/",
            "/home/admin/",
            "/home/AppUser/",
            "/home/MindXOM/",
            "/home/MEFEdge/",
            "/home/package/",
            "/home/HwHiAiUser/",
        )
        path = os.path.normpath(path)
        # 路径转为以/结尾
        path = path + "/" if path[:-1] != "/" else path
        path_in_white = False
        for item in white_paths:
            if len(path) <= len(item):
                continue
            if not path.startswith(item):
                continue
            path_in_white = True
            break
        if path == "/var/lib/docker/":
            path_in_white = True
        if not path_in_white:
            run_log.error("[%s] not in white path list", path)
            return False

        for item in black_paths:
            if path.startswith(item):
                run_log.error("[%s] is in black path list", path)
                return False

        # 挂载时检查的路径必定是不存在的，需要跳过不存在的目录层级，只检查存在的目录的属主是否root
        ret = FileCheck.check_path_is_root(path, skip_no_exists=True)
        if not ret:
            run_log.error("[%s] is invalid: %s", path, ret.error)
            return False

        return True

    @staticmethod
    def check_mount_path_is_subdirectory_of_mounted_path(mount_path):
        """
        功能描述：检查待挂载路径是否与已挂载路径间存在父子关系
        参数：@mount_path 待挂载路径
        返回值：bool
        """
        get_mounted_path_cmd_list = "%s -la | awk 'NR!=1 {print $6}'" % cmd_constants.OS_CMD_DF
        ret = ExecCmd.exec_cmd_use_pipe_symbol(get_mounted_path_cmd_list, wait=5)
        if ret[0] != 0:
            run_log.error("%s" % ret)
            return False

        mounted_path = ret[1].strip().split("\n")

        try:
            mounted_path.extend(nfs.local_dir for nfs in query_nfs_configs())
        except Exception as err:
            run_log.error("Query nfs config err. Catch %s", err.__class__.__name__)

        mount_path = os.path.join(mount_path, '')
        for path in mounted_path:
            # 排除根目录和空目录，以及系统挂载分区/opt
            if path == '/' or path == '' or path == '/opt':
                continue
            # 确保路径最后以"/"结尾，避免下一步判断失败
            path = os.path.join(path, '')
            if mount_path.startswith(path) or path.startswith(mount_path):
                run_log.error("There is a subdirectory relationship between Mount Path and Mounted path")
                return False

        return True

    @staticmethod
    def persist_mount_path(partition: str, path: Union[str, None], config_file: str) -> NoReturn:
        full_partition = f"/dev/{partition}"
        cfg_parser = configparser.ConfigParser()
        if path:
            run_log.info("add mount section %s" % partition)
            try:
                cfg_parser.read(config_file)
                cfg_parser.remove_section(partition)
                cfg_parser.add_section(partition)
                cfg_parser.set(partition, full_partition, path)
            except Exception as err:
                run_log.error("add mount section %s failed, %s", partition, err)
                return
        else:
            run_log.info("remove mount section %s", partition)
            try:
                cfg_parser.read(config_file)
                cfg_parser.remove_section(partition)
            except Exception as err:
                run_log.error("remove mount section %s failed, %s", partition, err)
                return

        Partition.write_config_file(config_file, cfg_parser)

    @staticmethod
    def umount_docker_path(mount_path):
        # 有容器则不取消挂载
        ret = ExecCmd.exec_cmd_get_output((shutil.which("docker"), "ps", "-qa"))
        if ret[0] != 0 or ret[1]:
            run_log.error("get docker container failed or containers have not been deleted")
            return False

        # 停止docker服务
        if ExecCmd.exec_cmd((cmd_constants.OS_CMD_SYSTEMCTL, "stop", "docker")) != 0:
            run_log.error("stop docker service failed")
            return False

        # 解挂docker目录及其下的所有子目录
        ret = ExecCmd.exec_cmd_get_output((cmd_constants.OS_CMD_UMOUNT, "-Rf", shlex.quote(mount_path)))
        if ret[0] != 0:
            run_log.error("umount %s failed: %s", mount_path, ret[1])
            return False

        return True

    @staticmethod
    def mount_docker_path():
        # 没有装MEF或者能力项关闭docker目录则按正常处理
        mef_info = MefInfo()
        cmd = (mef_info.run_sh, "restart")
        if not mef_info.allow_upgrade:
            # 启动docker服务
            cmd = (cmd_constants.OS_CMD_SYSTEMCTL, "restart", "docker")

        if ExecCmd.exec_cmd(cmd) != 0:
            run_log.error("restart docker service failed.")
            return False

        return True

    @staticmethod
    def _partition_id_check(partition_id: str) -> bool:
        return partition_id and AppCommonMethod.partition_id_check(partition_id)

    @staticmethod
    def _disk_generator() -> Iterable[Disk]:
        for dev_path in itertools.chain.from_iterable(Storage().get_storages().values()):
            try:
                yield DeviceLoader().load_disk(dev_path.as_posix())
            except Exception as err:
                msg = str(err) if isinstance(err, DevError) else f"catch {err.__class__.__name__}"
                run_log.warning("load block %s failed, %s, ignore.", dev_path.name, msg)

    @staticmethod
    def _part_is_primary(part: Part) -> bool:
        primary_partitions = const.A500_PRIMARY if SystemUtils().is_a500 else DeviceLoader.cfg.primary_partitions
        return all((part.disk_name == Env().boot_part.disk_name, int(part.part_num) in primary_partitions))

    def get_all_info(self, pat_name: str = None, pat_location: str = None) -> List[Union[int, str]]:
        """
        功能描述：通过磁盘名获取磁盘所有信息
        参数：pat_name 分区名 为None是获取资源列表
        参数：pat_location 设备的位置 为None是获取资源列表
        返回值：无
        异常描述：NA
        """
        if pat_name is None:
            self._refresh_cache()
            self.items = self.part_cache
            return [CommonMethods.OK, ]

        # pat_name传入空字符串，且pat_location不为空
        if not pat_name and pat_location:
            try:
                self.DeviceName = DeviceLoader().get_disk_path_by_location(pat_location)
            except Exception as err:
                msg = str(err) if isinstance(err, DevError) else f"catch {err.__class__.__name__}"
                run_log.error("get dev path of %s failed, %s", pat_location, msg)
                return [CommonMethods.NOT_EXIST, "input para invalid"]

            return [CommonMethods.OK, ]

        if pat_name not in self.part_cache or not self._partition_id_check(pat_name):
            return [CommonMethods.NOT_EXIST, 'input para invalid']

        try:
            part = DeviceLoader().load_part(pat_name)
        except Exception as err:
            msg = str(err) if isinstance(err, DevError) else f"catch {err.__class__.__name__}"
            run_log.error("get partition %s failed, %s.", pat_name, msg)
            return [CommonMethods.NOT_EXIST, f"{pat_name} not exists."]

        # 系统分区不支持查询
        if self._part_is_primary(part):
            return [CommonMethods.NOT_EXIST, "input para invalid"]

        try:
            self.Location = DeviceLoader().get_fmt_info(part.disk_path).location
        except Exception as err:
            msg = str(err) if isinstance(err, DevError) else f"catch {err.__class__.__name__}"
            run_log.error("get location of %s failed, %s", pat_location, msg)
            return [CommonMethods.NOT_EXIST, "input para invalid"]

        self.Name = pat_name
        self.FileSystem = part.fs_type
        self.CapacityBytes = str(part.size)
        self.FreeBytes = str(part.free())
        self.MountPath = part.mount_point
        # 应该是冗余字段，前端未使用，smartkit与FD未知
        self.Device = f"/redfish/v1/Systems/SimpleStorages/{part.storage_name.upper()}"
        self.DeviceName = Path(const.DEV_ROOT, part.disk_site_name).as_posix()
        self.Status = const.STATE_GOOD
        return [CommonMethods.OK, ]

    def post_request(self, request_dict: Union[dict, None]) -> List[Union[int, str]]:
        """创建分区"""
        if self.lock.locked():
            run_log.warning("post_request failed, partition lock is locked.")
            return [CommonMethods.ERROR, ERROR_MESSAGE.get("busy")]

        with self.lock:
            # 外部输入参数校验
            check_ret = CreatePatitionChecker().check(request_dict)
            if not check_ret.success:
                run_log.error("Create partition checker: %s.", check_ret.reason)
                return [CommonMethods.ERROR, ERROR_MESSAGE.get("param_wrong")]

            device = request_dict["Links"][0]["Device"]["@odata.id"]
            if os.path.basename(device) not in self.disk_cache:
                run_log.error("Illegal disk passed in, partition creation not supported.")
                return [CommonMethods.NOT_EXIST, ERROR_MESSAGE.get("disk_not_exist")]

            # A500 EMMC磁盘不支持创建分区
            if os.path.basename(device) == const.EMMC0:
                run_log.error("system partition not allow create disk partition.")
                return [CommonMethods.ERROR, ERROR_MESSAGE.get("param_wrong")]

            try:
                disk = DeviceLoader().load_disk(device)
            except Exception as err:
                msg = err if isinstance(err, DevError) else f"catch {err.__class__.__name__}."
                run_log.error("get disk %s failed, %s", device, msg)
                return [CommonMethods.NOT_EXIST, ERROR_MESSAGE.get("disk_not_exist")]

            # 检查磁盘容量
            capacity_gb = request_dict.get("CapacityBytes")
            number = request_dict.get("Number")
            avail_bytes = disk.free()
            run_log.info("disk %s(%s) remain space %s Bytes.", disk.name, disk.site_name, avail_bytes)
            if float(capacity_gb) * int(number) * const.KILO_BINARY_BYTE_RATE ** 3 > avail_bytes:
                run_log.error("Disk %s(%s) not enough space.", disk.name, disk.site_name)
                return [CommonMethods.ERROR, ERROR_MESSAGE.get("not_enough_space")]

            # 检查分区数是否超过阈值16
            pat_num = len(disk.children)
            run_log.info("The number of existing partitions is %s, and %s need to be added.", pat_num, number)
            if (pat_num + number) > const.MAX_PARTITION_NUM:
                run_log.error("The number of partitions exceeds the limit.")
                return [CommonMethods.ERROR, ERROR_MESSAGE.get("partition_num_error")]

            # 创建分区
            try:
                self._create_partitions(number, float(capacity_gb) * const.KILO_BINARY_BYTE_RATE ** 2, disk)
            except Exception as err:
                err_msg = err if isinstance(err, DevError) else f"create part failed, catch {err.__class__.__name__}."
                run_log.error(err_msg)
                return err.out() if isinstance(err, DevError) else CreatePartFailed.out()

            return [CommonMethods.OK, "OK"]

    def patch_request(self, request_dict: dict, partition_name=None) -> List[Union[int, str]]:
        """挂载与解挂分区"""
        if not self._partition_id_check(partition_name):
            return [CommonMethods.ERROR, ERROR_MESSAGE.get("partition_name_wrong")]

        if self.lock.locked():
            run_log.warning("patch_request failed, partition lock is locked.")
            return [CommonMethods.ERROR, ERROR_MESSAGE.get("busy")]

        with self.lock:
            user_name = request_dict.get("_User")
            request_ip = request_dict.get("_Xip")
            operator_check = CommonCheck.check_operator(user_name, request_ip)
            if not operator_check:
                run_log.error("The operator is illegal, %s", operator_check.error)
                return [CommonMethods.ERROR, ERROR_MESSAGE.get("operator_illegal")]

            if partition_name not in self.part_cache:
                run_log.error("Illegal part passed in, mount or umount part not supported.")
                return [CommonMethods.NOT_EXIST, f"{partition_name} not exists."]

            try:
                part = DeviceLoader().load_part(partition_name)
            except Exception as err:
                run_log.error("get partition %s failed, catch %s.", partition_name, err.__class__.__name__)
                return [CommonMethods.NOT_EXIST, f"{partition_name} not exists."]

            # 系统分区不支持挂载解挂
            if self._part_is_primary(part):
                return [CommonMethods.NOT_EXIST, "input para invalid"]

            mount_path = request_dict.get("MountPath")
            if mount_path:
                try:
                    self.mount_disk(mount_path, part)
                except Exception as err:
                    msg = err if isinstance(err, DevError) else f"mount part failed, catch {err.__class__.__name__}."
                    run_log.error(msg)
                    operate_log.info("[%s@%s] Mount the disk %s failed.", user_name, request_ip, partition_name)
                    return err.out() if isinstance(err, DevError) else MountFailed.out()

                operate_log.info("[%s@%s] Mount %s successfully.", user_name, request_ip, partition_name)
                return [CommonMethods.OK, "Mounted successfully"]

            try:
                self.unmount_disk(part)
            except Exception as err:
                msg = err if isinstance(err, DevError) else f"umount part failed, catch {err.__class__.__name__}."
                run_log.error(msg)
                operate_log.info("[%s@%s] Umount the disk %s failed.", user_name, request_ip, partition_name)
                return err.out() if isinstance(err, DevError) else UmountFailed.out()

            operate_log.info("[%s@%s] Umount %s successfully.", user_name, request_ip, partition_name)
            return [CommonMethods.OK, "Umount successfully"]

    def mount_disk(self, mount_path: str, part: Part):
        """挂载磁盘"""
        if part.mount_point:
            raise PartMounted("partition already mounted.")
        if not self.check_path_whitelist(mount_path) or not self.check_path_is_permitted(mount_path):
            raise PathNotInWhite("mount path not in white list.")
        if os.path.exists(mount_path):
            raise MountPathExisted("mount path already exists.")
        # 检查路径是否存在父子关系
        if not self.check_mount_path_is_subdirectory_of_mounted_path(mount_path):
            raise MountPathInvalid("There is a subdirectory relationship between Mount Path and Mounted path")
        os.makedirs(mount_path, 0o755)
        ret = ExecCmd.exec_cmd_get_output([cmd_constants.OS_CMD_MOUNT, part.path, shlex.quote(mount_path)], wait=60)
        if ret[0] != 0:
            self.remove_mount_path(mount_path)
            raise MountFailed(f"Mount the disk {part.name} failed.")
        if mount_path == "/var/lib/docker" and not self.mount_docker_path():
            raise MountFailed(f"Mount the disk {part.name} failed.")
        self.persist_mount_path(part.site_name, mount_path, self.PERSIST_CFG)
        FilePermission.set_path_permission(mount_path, 0o1755)

    def unmount_disk(self, part: Part):
        """卸载磁盘"""
        if not part.mount_point:
            raise PartNotMount("partition not mounted.")
        mount_path = part.mount_point
        run_log.info("mount path is %s", mount_path)
        if not self.check_path_whitelist(mount_path) or not self.check_path_is_permitted(mount_path):
            raise PathNotInWhite("mount path not in white list.")
        if mount_path == "/var/lib/docker":
            if not self.umount_docker_path(mount_path):
                raise UmountDockerFailed(f"umount {mount_path} failed.")
        else:
            if ExecCmd.exec_cmd((cmd_constants.OS_CMD_UMOUNT, shlex.quote(part.path))) != 0:
                raise UmountFailed(f"umount {part.name} failed.")
        self.persist_mount_path(part.site_name, None, self.PERSIST_CFG)
        self.remove_mount_path(mount_path)

    def delete_request(self, request: dict, partition_name: str = None) -> List[Union[int, str]]:
        """删除分区"""
        if not self._partition_id_check(partition_name):
            return [CommonMethods.ERROR, ERROR_MESSAGE.get("partition_name_wrong")]

        if self.lock.locked():
            run_log.warning("delete_request failed, partition lock is locked.")
            return [CommonMethods.ERROR, ERROR_MESSAGE.get("busy")]

        with self.lock:
            operator_check = CommonCheck.check_operator(request.get("_User"), request.get("_Xip"))
            if not operator_check:
                run_log.error("The operator is illegal, %s", operator_check.error)
                return [CommonMethods.ERROR, ERROR_MESSAGE.get("operator_illegal")]

            if partition_name not in self.part_cache:
                run_log.error("Illegal part passed in, delete part not supported.")
                return [CommonMethods.NOT_EXIST, f"{partition_name} not exists."]

            try:
                part = DeviceLoader().load_part(partition_name)
            except Exception as err:
                run_log.error("get partition %s failed, catch %s", partition_name, err.__class__.__name__)
                return [CommonMethods.NOT_EXIST, f"{partition_name} not exists."]

            # A500 EMMC磁盘不支持删除分区，主分区也不允许删除
            if any((part.disk_name == const.EMMC0, self._part_is_primary(part))):
                run_log.error("system partition not allow delete disk partition.")
                return [CommonMethods.ERROR, ERROR_MESSAGE.get("param_wrong")]

            if part.mount_point:
                try:
                    self.unmount_disk(part)
                except Exception as err:
                    msg = err if isinstance(err, DevError) else f"umount part failed, catch {err.__class__.__name__}."
                    run_log.error(msg)
                    return err.out() if isinstance(err, DevError) else UmountFailed.out()
            else:
                # 当分区挂载异常，删除挂载时，将配置文件中的分区信息也对应清除
                self.persist_mount_path(part.site_name, None, self.PERSIST_CFG)

            try:
                self._rm_part(part)
            except Exception as err:
                err_msg = err if isinstance(err, DevError) else f"del part failed, catch {err.__class__.__name__}."
                run_log.error(err_msg)
                return err.out() if isinstance(err, DevError) else DelPartFailed.out()

            return [CommonMethods.OK, "OK"]

    def _refresh_cache(self):
        self.part_cache.clear()
        self.disk_cache.clear()
        for disk in self._disk_generator():
            self.disk_cache.append(disk.site_name)
            for part in sorted(disk.children, key=lambda child: int(child.part_num)):
                if self._part_is_primary(part):
                    continue
                self.part_cache.append(part.site_name)

    def _set_disk_pt_type(self, disk: Disk):
        """设置磁盘格式为GPT"""
        if disk.pt_type != self.PT_TYPE and disk.children:
            raise ConvertGptError(f"{disk.pt_type} has partitions and is not in GPT format.")
        # 如果已经是GPT，无需转换
        if disk.pt_type == self.PT_TYPE:
            return

        self._sys_reload_disk(disk.path)
        with udev_monitor(partial(terminate_by_gpt, disk=disk), self.WAIT):
            status, info = self.exec_cmd((
                cmd_constants.OS_CMD_PARTED, "-s", shlex.quote(disk.path), "mklabel", self.PT_TYPE
            ))
            if status != 0:
                raise ConvertGptError(info)
        # 监控没有异常则说明格式转换成功
        run_log.info("convert gpt success.")

    def _set_part_fs_type(self, part: Part):
        status, info = self.exec_cmd((cmd_constants.OS_CMD_MKE2FS, "-t", self.FS_TYPE, shlex.quote(part.path)))
        if status != 0:
            raise DevError(f"mke2fs failed, {info}.")
        run_log.info("mke2fs success.")

    def _sys_reload_disk(self, dev_name: str):
        """告知内核重新加载磁盘信息"""
        try:
            with udev_monitor(partial(terminate_by_path, path=dev_name), timeout=self.WAIT):
                status, msg = self.exec_cmd((cmd_constants.OS_CMD_PARTPROBE, shlex.quote(dev_name)), wait=self.WAIT)
                if status != 0:
                    raise DevError("exec partprobe failed.")
        except Exception as err:
            reason = f"because of {err}" if isinstance(err, DevError) else f"catch {err.__class__.__name__}."
            run_log.warning("reload %s failed, %s", dev_name, reason)

    def _create_partitions(self, number: int, size_kib: float, disk: Disk):
        self._set_disk_pt_type(disk)
        for _ in range(number):
            malloc_start, malloc_end = disk.offset_for_new_part(size_kib, Env().boot_part)
            malloc_start, malloc_end = 2048 * int(malloc_start / 2048) + 2048, 2048 * int(malloc_end / 2048) + 2048
            run_log.info([malloc_start, malloc_end])
            mk_part = (
                cmd_constants.OS_CMD_PARTED, "-s", shlex.quote(disk.path), "mkpart", "primary", "ext2",
                f"{malloc_start}kiB", f"{malloc_end}kiB"
            )
            with udev_monitor(partial(terminate_while_add_part, disk=disk), timeout=self.WAIT,
                              subsystem=SubSystem.PART):
                status, info = self.exec_cmd(mk_part)
                if status != 0:
                    raise CreatePartFailed(info)
            # 创建成功后，重新加载磁盘
            old_parts = [part.name for part in disk.children]
            self._sys_reload_disk(disk.path)
            disk = DeviceLoader().load_disk(disk.path)
            new_parts = [part for part in disk.children if part.name not in old_parts]
            if len(new_parts) != 1:
                self._try_rm_partitions(*new_parts)
                raise CreatePartFailed(f"incorrect number of new partitions. {len(new_parts)} partitions found.")
            new_part = new_parts[0]
            try:
                self._set_part_fs_type(new_part)
            except DevError as err:
                self._try_rm_partitions(new_part)
                raise CreatePartFailed("mke2fs to ext4 failed.") from err
            run_log.info("create partition %s success.", new_part.path)
            self.DeviceName = new_part.site_name
            run_log.info("Create partition %s success, capacity is %s Bytes.", new_part.site_name, new_part.size)
            # 执行完单个分区创建后，强制刷新下分区信息，防止偶现的设置了fs_type后，分区fs_type不显示问题
            self._sys_reload_disk(disk.path)

    def _exec_rm_part(self, part: Part):
        terminate_func = partial(terminate_by_path, path=part.path, action=DevAction.RM)
        with udev_monitor(terminate_func, subsystem=SubSystem.PART, timeout=self.WAIT):
            rm_part = cmd_constants.OS_CMD_PARTED, "-s", shlex.quote(part.disk_path), "rm", part.part_num
            status, info = self.exec_cmd(rm_part)
            if status != 0:
                raise DelPartFailed(f"rm partition {part.path} failed.")
        # 删除完成后，让系统重新加载一下磁盘
        self._sys_reload_disk(part.disk_path)

        run_log.info("rm partition %s success.", part.path)

    def _try_rm_partitions(self, *parts: Part):
        for part in parts:
            try:
                self._exec_rm_part(part)
            except Exception as err:
                run_log.error("try del %s failed, catch %s.", part.name, err.__class__.__name__)

    def _rm_part(self, part: Part):
        try:
            self._set_part_fs_type(part)
        except DevError as err:
            raise DelPartFailed(str(err)) from err
        self._exec_rm_part(part)
