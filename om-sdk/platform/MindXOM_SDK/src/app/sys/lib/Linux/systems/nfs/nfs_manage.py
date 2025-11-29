# !/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os
import shlex
import threading
import time
from contextlib import contextmanager
from functools import partial
from pathlib import Path
from typing import Callable, Iterable, List, Union, Optional

from bin.environ import Env
from common.common_methods import CommonMethods
from common.file_utils import FileAttribute, FileUtils, FilePermission
from common.file_utils import FileCreate
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd
from common.utils.system_utils import SystemUtils
from common.utils.timer import RepeatingTimer
from lib.Linux.systems.nfs import errors
from lib.Linux.systems.nfs.cfg_mgr import (query_nfs_configs, del_nfs_config_by_mount_path, save_nfs_config,
                                           mount_path_already_exists)
from lib.Linux.systems.nfs.models import NfsCfg, get_nfs_status, nfs_server_valid
from lib.Linux.systems.nfs.schemas import NfsReq, Operation, NfsStatus
from lib.Linux.systems.nic.config_web_ip import NginxConfig


class NfsManage:
    NFS_MANAGER_LOCK = threading.Lock()
    # 解挂会影响NFS状态检测，最差正在进行状态检测且对应NFS服务端关闭，阻塞10s、尝试解挂1s，正常检测一个状态不会很长
    UMOUNT_WAIT_SECOND = 12
    BACKUP_REGION = "/run/backup_region"
    MOUNT_OPTIONS = "-t nfs4 -o noexec,nosuid,proto=tcp,rsize=1048576,wsize=1048576,soft,intr,retry=1,retrans=3"
    MONITOR_PERIOD = 30
    TM_OUT = 10

    umount_lock = threading.Lock()
    exec_cmd: Callable[[Iterable, int], List[Union[int, str]]] = partial(ExecCmd.exec_cmd_get_output)
    request: Optional[NfsReq] = None
    status_cache = NfsStatus()

    def __init__(self):
        """
        功能描述：初始化函数
        参数：
        返回值：无
        异常描述：NA
        """
        self.nfsList = []

    @contextmanager
    def acquire_umount_lock_with_timeout(self):
        acquired = self.umount_lock.acquire(timeout=self.UMOUNT_WAIT_SECOND)
        try:
            yield acquired
        finally:
            if acquired:
                self.umount_lock.release()

    def get_all_info(self):
        """
        功能描述：通过Nfs挂载信息
        参数：NA
        返回值：无
        异常描述：NA
        修改记录：2019-8-17 创建
        """

        self.nfsList = []
        try:
            self.nfsList = [cfg.to_dict() for cfg in query_nfs_configs()]
        except Exception as err:
            run_log.error(str(err))
        if not self.nfsList:
            run_log.info("nfs cfg is null.")

    def post_request(self, request_dict=None):
        """
        功能描述：解析post请求
        参数：
        返回值：无
        异常描述：NA
        修改记录：2019-8-17 创建
        """
        if NfsManage.NFS_MANAGER_LOCK.locked():
            run_log.warning("Nfs modify is busy")
            return errors.OperateBusy.out()

        with NfsManage.NFS_MANAGER_LOCK:
            try:
                NfsManage.request = NfsReq.from_dict(request_dict)
            except errors.MountError as err:
                run_log.error(err)
                return err.out()
            except Exception as err:
                run_log.error("Check nfs catch %s", err.__class__.__name__)
                return errors.OperateFailed.out()

            if self.request.operate == Operation.MOUNT.value:
                return self.mount()

            with self.acquire_umount_lock_with_timeout() as acquired:
                if acquired:
                    return self.umount()

                run_log.warning("Performing state detection for NFS.")
                return errors.OperateBusy.out()

    def get_mount_options(self):
        try:
            _, nginx_listen_ipv4 = NginxConfig.get_nginx_listen_ipv4()
        except Exception as err:
            run_log.error("get nginx listen ipv4 failed. %s", err)
            raise errors.OperateFailed("get nginx listen ipv4 failed!")

        return f"{self.MOUNT_OPTIONS},clientaddr={nginx_listen_ipv4}"

    def umount(self):
        try:
            self.umount_nfs(self.request.mount_path)
        except errors.MountError as err:
            run_log.error("Umount nfs err, %s", err)
            return err.out()
        except Exception as err:
            run_log.error("Umount nfs dir catch %s", err.__class__.__name__)
            return errors.OperateFailed.out()

        cfg = NfsCfg(self.request.server_ip, self.request.server_dir, self.request.mount_path, self.request.fs_type)
        try:
            del_nfs_config_by_mount_path(self.request.mount_path)
        except Exception as err:
            run_log.error("del nfs config err, catch %s", err.__class__.__name__)
            self.try_mount_nfs(cfg)
            return errors.OperateFailed.out()

        self.status_cache.delete_status(cfg.status_key_string)
        self.try_save_status()

        return [CommonMethods.OK, "Umount successfully."]

    def mount(self):
        cfg = NfsCfg(self.request.server_ip, self.request.server_dir, self.request.mount_path, self.request.fs_type)
        try:
            self.mount_nfs(cfg)
        except Exception as err:
            try_remove_path(cfg.local_dir)
            if isinstance(err, errors.MountError):
                run_log.error("Mount nfs failed, %s", err)
                return err.out()

            run_log.error("Mount nfs failed, catch %s", err.__class__.__name__)
            return errors.OperateFailed.out()

        # 需在session保存成功前，先取对应实例的属性供设状态，否则会触发Instance NfsCfg is not bound to a Session
        status_key = cfg.status_key_string
        try:
            save_nfs_config(cfg)
        except Exception as err:
            run_log.error("save nfs config failed, catch %s", err.__class__.__name__)
            self.umount_nfs(cfg.local_dir)
            return errors.OperateFailed.out()

        self.status_cache.set_status_ok(status_key)
        self.try_save_status()

        return [CommonMethods.OK, "Mount successfully."]

    def umount_nfs(self, mount_path):
        """卸载"""
        run_log.info("nfs umount operation! %s", mount_path)
        if not umount_path(mount_path, is_nfs=True):
            raise errors.OperateFailed("Umount NFS failed.")
        try_remove_path(mount_path)

        # 500场景下，对区的挂载点解挂时需要清理；其他场景可能不存在对区。
        if not SystemUtils().is_a500:
            return

        try:
            self.rm_region_backup_nfs_dir(mount_path)
        except Exception as err:
            run_log.error("rm remain dir catch %s", err.__class__.__name__)
        finally:
            # 如果BACKUP_REGION已挂载，需解卦删除
            if umount_path(self.BACKUP_REGION):
                try_remove_path(self.BACKUP_REGION)

    def rm_region_backup_nfs_dir(self, local_dir: str):
        """
        假如当前在p2分区，通过web正常创建nfs共享目录test，重启系统切换到p3分区解挂共享目录，解挂成功。
        再切回p2时，test目录依旧存在，需要清理。
        尝试清理，遇到异常则忽略清理
        """
        if not FileCreate.create_dir(self.BACKUP_REGION, 0o700):
            run_log.warning("Read cmdline or create backup region failed.")
            return

        region_backup = Env().back_partition.as_posix()
        status, msg = self.exec_cmd((cmd_constants.OS_CMD_MOUNT, region_backup, self.BACKUP_REGION))
        if status != 0:
            run_log.warning("mount backup region failed.")
            return

        run_log.info("mount backup region success")
        remain_dir = os.path.join(self.BACKUP_REGION, *local_dir.split(os.sep))
        if not os.path.exists(remain_dir):
            # 没有残留，直接返回
            run_log.info("remain dir not exists.")
            return

        # 尝试chattr -i 并删除残留目录
        run_log.info("try to remove remain dir")
        FileAttribute.set_path_immutable_attr(remain_dir, False, False)
        try_remove_path(remain_dir)

    def mount_nfs(self, cfg: NfsCfg):
        """挂载"""
        if not FileCreate.create_dir(cfg.local_dir, 0o400):
            raise errors.OperateFailed("Create mount point failed")
        if not FileAttribute.set_path_immutable_attr(cfg.local_dir, True, False):
            raise errors.OperateFailed("Set mount path chatter error!")
        run_log.info("Nfs mount operation! %s %s %s", cfg.files_system, cfg.local_dir, cfg.fs_type)
        status, msg = self.exec_cmd((
            "timeout", str(self.TM_OUT), cmd_constants.OS_CMD_MOUNT, *self.get_mount_options().split(),
            shlex.quote(cfg.files_system), shlex.quote(cfg.local_dir)
        ))
        if status == -1000:
            raise errors.TimeOut("Mount time out, network or nfs server error.")
        elif status != 0:
            raise errors.OperateFailed("Mount nfs failed.")
        run_log.info("Mount success!")
        FilePermission.set_path_permission(cfg.local_dir, 0o1755)

    def try_mount_nfs(self, cfg):
        try:
            self.mount_nfs(cfg)
        except Exception as err:
            msg = str(err) if isinstance(err, errors.MountError) else f"catch {err.__class__.__name__}"
            run_log.error("Monitor:Mount %s failed, %s", cfg.local_dir, msg)
            FileAttribute.set_path_immutable_attr(cfg.local_dir, False, False)
            try_remove_path(cfg.local_dir)

    def monitor(self):
        """定时任务，未挂载则挂载，挂载异常则尝试挂载"""
        if self.NFS_MANAGER_LOCK.locked():
            run_log.warning("NFS modify is busy")
            return

        # list触发生成器执行一遍，跳出session，避免嵌套session
        for cfg in list(query_nfs_configs(expunge=True)):
            # 睡一会儿提升解挂获取锁的概率
            time.sleep(1)
            with self.umount_lock:
                # 当检测开始后，如果解挂请求成功会将数据库记录删除，触发状态检测前应当再次判断挂载点是否存在
                if not mount_path_already_exists(cfg.local_dir):
                    continue

                server_valid = nfs_server_valid(cfg.local_dir)
                if server_valid and get_nfs_status(cfg):
                    self.status_cache.set_status_ok(cfg.status_key_string)
                    continue

                self.status_cache.set_status_error(cfg.status_key_string)
                # 尝试解挂
                if umount_path(cfg.local_dir, is_nfs=True):
                    try_remove_path(cfg.local_dir)

                # 尝试挂载
                self.try_mount_nfs(cfg)
                if get_nfs_status(cfg):
                    run_log.info("Try mount %s success.", cfg.local_dir)
                else:
                    run_log.warning("Try mount %s failed.", cfg.local_dir)

        self.try_save_status()

    def start_monitor(self):
        RepeatingTimer(self.MONITOR_PERIOD, self._monitor).start()

    def uninstall(self):
        """卸载时先尝试将所有NFS挂载点解挂"""
        for cfg in query_nfs_configs():
            try:
                self.umount_nfs(cfg.local_dir)
            except Exception as err:
                run_log.error("umount %s catch %s", cfg.local_dir, err.__class__.__name__)

    def try_save_status(self):
        try:
            self.status_cache.save()
        except Exception as err:
            run_log.warning("save nfs status error, catch %s", err.__class__.__name__)

    def _monitor(self):
        try:
            self.monitor()
        except Exception as err:
            run_log.warning("Execute nfs monitor failed, because of %s", err)


def umount_path(mount_point: str, is_nfs=False) -> bool:
    if not is_nfs and not Path(mount_point).is_mount():
        return False

    if not is_nfs:
        return ExecCmd.exec_cmd((cmd_constants.OS_CMD_UMOUNT, "-l", mount_point)) == 0

    # 如果是nfs解挂，调用umount 1s后判断挂载点是否存在，防止服务端关闭后执行umount卡死
    ExecCmd.exec_cmd(("timeout", "1", cmd_constants.OS_CMD_UMOUNT, "-l", mount_point))
    # 如果mount | grep mount_point | grep nfs4不存在则说明挂载点已解挂，该命令不会因nfs服务端关闭阻塞
    return not ExecCmd.exec_cmd_use_pipe_symbol(f"{cmd_constants.OS_CMD_MOUNT} | grep ' {mount_point} ' | grep nfs4")[1]


def try_remove_path(path: str):
    # 删除前先尝试chattr -i 属性
    FileAttribute.set_path_immutable_attr(path, False, False)
    try:
        FileUtils.delete_full_dir(path)
    except Exception as err:
        run_log.warning("Try remove path failed, catch %s", err.__class__.__name__)
