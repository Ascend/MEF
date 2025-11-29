# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
import os
from collections import namedtuple
from pathlib import Path

import pytest
from pytest_mock import MockFixture

from bin.environ import Env
from common.file_utils import FileCreate, FileAttribute, FileUtils
from common.utils.exec_cmd import ExecCmd
from common.utils.system_utils import SystemUtils
from lib.Linux.systems.nfs.errors import MountError, OperateFailed, TimeOut, ParmaError
from lib.Linux.systems.nfs.models import NfsCfg
from lib.Linux.systems.nfs.nfs_manage import NfsManage, umount_path, try_remove_path
from lib.Linux.systems.nic.config_web_ip import NginxConfig
from test_mqtt_api.get_log_info import GetLogInfo

getLog = GetLogInfo()

SERVER_IP = "10.10.10.10"
HOME_DIR = "/home"
NFS_TYPE = "nfs4"


class TestNfsManage:
    use_cases = {
        "test_get_all_info": {
            "none": ("nfs cfg is null.", []),
            "exception": ("ERROR", ["ada"]),
        },
        "test_post_request": {
            "locked": ([400, [100028, 'The operation is busy.']], None, True, None, None),
            "none_data": ([400, [100024, "Parameter is invalid."]], None, False, None, None),
            "exception": ([400, [110505, "Operate NFS failed."]], "exception", False, None, None),
            "mount": ([400, [110505, "Operate NFS failed."]],
                      {
                          "ServerIP": "51.38.66.252",
                          "ServerDir": "/mnt/nfs",
                          "MountPath": "/opt/mount",
                          "FileSystem": NFS_TYPE,
                          "Type": "mount"
                      },
                      False, None, None),
            "umount": ([400, [110507, 'Unmount path does not exist.']],
                       {
                           "ServerIP": "51.38.66.252",
                           "ServerDir": "/mnt/nfs",
                           "MountPath": "",
                           "FileSystem": NFS_TYPE,
                           "Type": "umount"
                       },
                       False, None, None, ),
        },
        "test_get_mount_options": {
            "normal": ("-t nfs4 -o noexec,nosuid,proto=tcp,rsize=1048576,wsize=1048576,soft,"
                       "intr,retry=1,retrans=3,clientaddr=10.10.10.10", ("", SERVER_IP)),
            "exception": ("get nginx listen ipv4 failed.", "test")
        },
        "test_umount_nfs": {
            "exception": (ParmaError(), [False, False], None),
            "not_a500": (None, [True, True], False),
            "normal": (None, [True, True], True),
        },
        "test_rm_region_backup_nfs_dir": {
            "create_dir_failed": ("Read cmdline or create backup region failed.", False, [None, None], None),
            "exec_cmd_failed": ("mount backup region failed.", True, [1, "error"], None),
            "not_exists": ("remain dir not exists.", True, [0, "OK"], [False, True]),
            "normal": ("try to remove remain dir", True, [0, "OK"], [True, True]),
        },
        "test_mount_nfs": {
            "create_dir_failed": (OperateFailed(), False, None, None, None),
            "set_attr_failed": (OperateFailed(), True, False, None, None),
            "overtime": (TimeOut(), True, True, [-1000, None], None),
            "mount_failed": (OperateFailed(), True, True, [1, None], None),
            "succeed": ("Mount success!", True, True, [0, None], None),
        },
        "test_try_mount_nfs": {
            "succeed": ("", [True]),
            "exception": ("Monitor:Mount ", MountError),
        },
        "test_umount_path": {
            "not_nfs_and_not_mount": (False, False, False, [None, None], None),
            "not_nfs_and_umount_failed": (False, False, True, [-1, None], None),
            "not_nfs_and_umount_succeed": (True, False, True, [0, None], None),
            "check_mount_failed": (True, True, True, [0, None], [None, False]),
            "check_mount_succeed": (False, True, True, [0, None], [None, True]),
        },
        "test_try_remove_path": {
            "normal": ("", [True]),
            "exception": ("Try remove path failed, catch", Exception),
        },
    }

    TryRemovePathCase = namedtuple("TryRemovePathCase", "expect, delete_full_dir")
    UmountPathCase = namedtuple("UmountPathCase", "expect, is_nfs, is_mount, exec_cmd, exec_cmd_use_pipe_symbol")
    TryMountNfsCase = namedtuple("TryMountNfsCase", "expect, mount_nfs")
    MountNfsCase = namedtuple("MountNfsCase", "expect, create_dir, set_attr, exec_cmd, set_permission")
    RmNfsDirCase = namedtuple("RmNfsDirCase", "expect, create_dir, exec_cmd, exists")
    UmountNfsCase = namedtuple("UmountNfsCase", "expect, umount_path, is_a500")
    GetMountOptionsCase = namedtuple("GetMountOptionsCase", "expect, get_ipv4")
    PostRequestCase = namedtuple("PostRequestCase", "expect, request_dict, lock, mount, umount")
    GetAllInfoCase = namedtuple("GetAllInfoCase", "expect, query_nfs_configs")

    @staticmethod
    def test_try_remove_path(mocker: MockFixture, model: TryRemovePathCase):
        mocker.patch.object(FileAttribute, "set_path_immutable_attr", return_value=None)
        mocker.patch.object(FileUtils, "delete_full_dir", side_effect=model.delete_full_dir)
        getLog.clear_log()
        try_remove_path(HOME_DIR)
        assert model.expect in getLog.get_log()

    @staticmethod
    def test_umount_path(mocker: MockFixture, model: UmountPathCase):
        mocker.patch.object(Path, "is_mount", return_value=model.is_mount)
        mocker.patch.object(ExecCmd, "exec_cmd", side_effect=model.exec_cmd)
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol", return_value=model.exec_cmd_use_pipe_symbol)
        assert umount_path(mount_point=HOME_DIR, is_nfs=model.is_nfs) == model.expect

    @staticmethod
    def test_try_mount_nfs(mocker: MockFixture, model: TryMountNfsCase):
        mocker.patch.object(NfsManage, "mount_nfs", side_effect=model.mount_nfs)
        mocker.patch.object(FileAttribute, "set_path_immutable_attr", return_value=None)
        mocker.patch("lib.Linux.systems.nfs.nfs_manage.try_remove_path", return_value=None)
        getLog.clear_log()
        NfsManage().try_mount_nfs(cfg=NfsCfg(
            server_ip=SERVER_IP, server_dir=HOME_DIR, local_dir=HOME_DIR, fs_type=NFS_TYPE))
        assert model.expect in getLog.get_log()

    @staticmethod
    def test_mount_nfs(mocker: MockFixture, model: MountNfsCase):
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        mocker.patch.object(FileAttribute, "set_path_immutable_attr", return_value=model.set_attr)
        mocker.patch.object(NfsManage, "exec_cmd", return_value=model.exec_cmd)
        mocker.patch.object(NfsManage, "get_mount_options", return_value="-t nfs4 -o noexec,nosuid,"
                                                                         "proto=tcp,rsize=1048576,wsize=1048576,"
                                                                         "soft,intr,retry=1,retrans=3,"
                                                                         "clientaddr=10.10.10.10")
        if isinstance(model.expect, OperateFailed):
            with pytest.raises(OperateFailed):
                NfsManage().mount_nfs(cfg=NfsCfg(
                    server_ip=SERVER_IP, server_dir=HOME_DIR, local_dir=HOME_DIR, fs_type=NFS_TYPE))
        elif isinstance(model.expect, TimeOut):
            with pytest.raises(TimeOut):
                NfsManage().mount_nfs(cfg=NfsCfg(
                    server_ip=SERVER_IP, server_dir=HOME_DIR, local_dir=HOME_DIR, fs_type=NFS_TYPE))
        else:
            getLog.clear_log()
            NfsManage().mount_nfs(cfg=NfsCfg(
                server_ip=SERVER_IP, server_dir=HOME_DIR, local_dir=HOME_DIR, fs_type=NFS_TYPE))
            assert model.expect in getLog.get_log()

    @staticmethod
    def test_rm_region_backup_nfs_dir(mocker: MockFixture, model: RmNfsDirCase):
        getLog.clear_log()
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        mocker.patch.object(NfsManage, "exec_cmd", return_value=model.exec_cmd)
        mocker.patch.object(Env, "back_partition", return_value=Path("/dev", "".join(("test", "3"))))
        mocker.patch.object(os.path, "exists", side_effect=model.exists)
        mocker.patch.object(FileAttribute, "set_path_immutable_attr", return_value=None)
        mocker.patch("lib.Linux.systems.nfs.nfs_manage.try_remove_path", return_value=None)
        NfsManage().rm_region_backup_nfs_dir("/test/test")
        assert model.expect in getLog.get_log()

    @staticmethod
    def test_umount_nfs(mocker: MockFixture, model: UmountNfsCase):
        mocker.patch("lib.Linux.systems.nfs.nfs_manage.umount_path", side_effect=model.umount_path)
        mocker.patch.object(SystemUtils, "is_a500", return_value=model.is_a500)
        if isinstance(model.expect, Exception):
            with pytest.raises(OperateFailed):
                NfsManage().umount_nfs("test")
        else:
            assert NfsManage().umount_nfs("test") == model.expect

    @staticmethod
    def test_get_mount_options(mocker: MockFixture, model: GetMountOptionsCase):
        mocker.patch.object(NginxConfig, "get_nginx_listen_ipv4", return_value=model.get_ipv4)
        if not isinstance(model.get_ipv4, tuple):
            with pytest.raises(OperateFailed):
                NfsManage().get_mount_options()
        else:
            assert NfsManage().get_mount_options() == model.expect

    @staticmethod
    def test_post_request(mocker: MockFixture, model: PostRequestCase):
        mocker.patch.object(NfsManage, "NFS_MANAGER_LOCK").locked.return_value = model.lock
        mocker.patch.object(NfsManage, "mount", return_value=model.mount)
        mocker.patch.object(NfsManage, "umount", return_value=model.umount)
        assert NfsManage().post_request(model.request_dict) == model.expect

    @staticmethod
    def test_get_all_info(mocker: MockFixture, model: GetAllInfoCase):
        getLog.clear_log()
        mocker.patch("lib.Linux.systems.nfs.nfs_manage.query_nfs_configs", side_effect=model.query_nfs_configs)
        NfsManage().get_all_info()
        assert model.expect in getLog.get_log()
