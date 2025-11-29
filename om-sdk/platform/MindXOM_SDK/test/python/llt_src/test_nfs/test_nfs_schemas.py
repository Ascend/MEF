# Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import os
import pytest
from pytest_mock import MockerFixture

from common.checkers import LocalIpChecker
from common.file_utils import FileCheck
from common.file_utils import FileCreate
from lib.Linux.systems.disk.partition import Partition
from lib.Linux.systems.nfs.schemas import Operation
from lib.Linux.systems.nfs.schemas import NfsReq
from lib.Linux.systems.nfs.schemas import NfsStatus
from lib.Linux.systems.nfs.schemas import MountValidator
from lib.Linux.systems.nfs.schemas import NFS_MAX_CFG_NUM
from lib.Linux.systems.nfs import schemas
from lib.Linux.systems.nfs import cfg_mgr


def test_operation():
    assert Operation.MOUNT.value == "mount" and Operation.UMOUNT.value == "umount"


class TestNfsStatus:
    @staticmethod
    def test_set_nfs_status():
        status_key, status = "test ok", "ok"
        nfs_status = NfsStatus()
        nfs_status.set_nfs_status(status_key, status)
        assert nfs_status.status_cache[status_key] == status

    @staticmethod
    def test_set_status_ok():
        status_key = "test ok"
        nfs_status = NfsStatus()
        nfs_status.set_status_ok(status_key)
        assert nfs_status.status_cache[status_key] == nfs_status.OK

    @staticmethod
    def test_set_status_error():
        status_key = "test ok"
        nfs_status = NfsStatus()
        nfs_status.set_status_error(status_key)
        assert nfs_status.status_cache[status_key] == nfs_status.ERROR

    @staticmethod
    def test_delete_status():
        status_key = "test ok"
        nfs_status = NfsStatus()
        nfs_status.set_status_ok(status_key)
        assert nfs_status.status_cache[status_key] == nfs_status.OK
        nfs_status.delete_status(status_key)
        assert status_key not in nfs_status.status_cache

    @staticmethod
    def test_save_with_not_exist_and_not_create(mocker: MockerFixture):
        mocker.patch.object(os.path, "exists", return_value=False)
        mocker.patch.object(FileCreate, "create_dir", return_value=False)
        with pytest.raises(Exception) as exception_info:
            NfsStatus().save()
            assert "create status path error." in str(exception_info.value)

    @staticmethod
    def test_save_with_check_failed(mocker: MockerFixture):
        mocker.patch.object(os.path, "exists", return_value=True)
        mocker.patch.object(FileCheck, "check_input_path_valid", return_value=False)
        with pytest.raises(Exception) as exception_info:
            NfsStatus().save()
            assert "status file invalid." in str(exception_info.value)


class TestNfsReq:
    @staticmethod
    def test_check_input_path_with_none_path():
        path = None
        assert NfsReq.check_input_path(path)

    @staticmethod
    def test_check_input_path_with_path_too_len():
        path = "invalid path" * 40
        assert not NfsReq.check_input_path(path)

    @staticmethod
    def test_check_input_path_with_path_is_invalid():
        path = ".."
        assert not NfsReq.check_input_path(path)

    @staticmethod
    def test_check_input_path_ok():
        path = "abc"
        assert NfsReq.check_input_path(path)

    @staticmethod
    def test_from_dict():
        data = dict()
        with pytest.raises(Exception) as exception_info:
            NfsReq.from_dict(data)
            assert "Parameter is null" in str(exception_info.value)

    @staticmethod
    def test_valid_operate():
        with pytest.raises(Exception) as exception_info:
            nfs_req = NfsReq(operate="", server_ip="", server_dir="", fs_type="", mount_path="")
            assert "Unknown nfs operation type" in str(exception_info.value)

    @staticmethod
    def test_valid_server_ip():
        with pytest.raises(Exception) as exception_info:
            nfs_req = NfsReq(operate="mount", server_ip="", server_dir="", fs_type="", mount_path="")
            assert "Invalid server ip" in str(exception_info.value)

    @staticmethod
    def test_valid_server_ip_checker_failed(mocker: MockerFixture):
        mocker.patch.object(LocalIpChecker, "check_dict", return_value=False)
        with pytest.raises(Exception) as exception_info:
            nfs_req = NfsReq(operate="mount", server_ip="1.1.1.1", server_dir="", fs_type="", mount_path="")
            assert "Invalid server ip" in str(exception_info.value)

    @staticmethod
    def test_valid_server_dir(mocker: MockerFixture):
        mocker.patch.object(LocalIpChecker, "check_dict", return_value=True)
        with pytest.raises(Exception) as exception_info:
            nfs_req = NfsReq(operate="mount", server_ip="1.1.1.1", server_dir="", fs_type="", mount_path="")
            assert "Server share path invalid" in str(exception_info.value)

    @staticmethod
    def test_valid_fs_type(mocker: MockerFixture):
        mocker.patch.object(LocalIpChecker, "check_dict", return_value=True)
        with pytest.raises(Exception) as exception_info:
            nfs_req = NfsReq(operate="mount", server_ip="1.1.1.1", server_dir="server", fs_type="", mount_path="")
            assert "Invalid file system type" in str(exception_info.value)


class TestMountValidator:
    @staticmethod
    def test_init():
        path = "abc"
        assert MountValidator(path).path == path

    @staticmethod
    def test_not_exceeds_limit(mocker: MockerFixture):
        mount_validator = MountValidator("")
        mocker.patch.object(schemas, "get_nfs_config_count", return_value=NFS_MAX_CFG_NUM)
        with pytest.raises(Exception) as exception_info:
            mount_validator.not_exceeds_limit()
            assert "NFS configuration exceeds limit." in str(exception_info.value)

    @staticmethod
    def test_not_exists_with_path_exist(mocker: MockerFixture):
        mount_validator = MountValidator("")
        mocker.patch.object(os.path, "exists", return_value=True)
        with pytest.raises(Exception) as exception_info:
            mount_validator.not_exists()
            assert "Mount path already exists." in str(exception_info.value)

    @staticmethod
    def test_not_exists_with_mount_path_exist(mocker: MockerFixture):
        mount_validator = MountValidator("")
        mocker.patch.object(os.path, "exists", return_value=False)
        mocker.patch.object(cfg_mgr, "mount_path_already_exists", return_value=True)
        with pytest.raises(Exception) as exception_info:
            mount_validator.not_exists()
            assert "Mount path already in configs." in str(exception_info.value)

    @staticmethod
    def test_is_not_subdir(mocker: MockerFixture):
        mount_validator = MountValidator("")
        mocker.patch.object(Partition, "check_mount_path_is_subdirectory_of_mounted_path", return_value=False)
        with pytest.raises(Exception) as exception_info:
            mount_validator.is_not_subdir()
            assert "a subdirectory relationship between Mount Path and Mounted path" in str(exception_info.value)


def test_valid_path_with_empty_path():
    with pytest.raises(Exception) as exception_info:
        schemas.valid_path("")
        assert "Mount path is null" in str(exception_info.value)


def test_valid_path_with_check_failed(mocker: MockerFixture):
    mocker.patch.object(FileCheck, "check_input_path_valid", return_value=False)
    with pytest.raises(Exception) as exception_info:
        schemas.valid_path("abc")
        assert "NFS Mount path invalid" in str(exception_info.value)


def test_is_whitelist_and_permitted_with_invalid_path():
    with pytest.raises(Exception) as exception_info:
        schemas.is_whitelist_and_permitted("/var/lib/docker")
        assert "Mount path is not in whitelist." in str(exception_info.value)


def test_is_whitelist_and_permitted_with_check_failed(mocker: MockerFixture):
    mocker.patch.object(Partition, "check_path_whitelist", return_value=False)
    with pytest.raises(Exception) as exception_info:
        schemas.is_whitelist_and_permitted("/var/lib/docker")
        assert "Mount path is not in whitelist." in str(exception_info.value)





