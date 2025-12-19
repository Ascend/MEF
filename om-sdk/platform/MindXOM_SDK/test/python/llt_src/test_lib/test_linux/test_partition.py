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
import json
from pathlib import Path
from typing import NamedTuple, Optional, Iterable, List, Tuple, Type, Any

import pytest
from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.utils.exec_cmd import ExecCmd
from lib.Linux.systems.disk.device_loader import DeviceLoader
from lib.Linux.systems.disk.errors import (PartMounted, PathNotInWhite, MountPathInvalid, MountFailed, DevError,
                                           PartNotMount, UmountDockerFailed, UmountFailed, MountPathExisted)
from lib.Linux.systems.disk.partition import Partition, ERROR_MESSAGE
from common.common_methods import CommonMethods


class PartIdCheckModel(NamedTuple):
    expect: bool
    part_id: Optional[str]


class MockPart(NamedTuple):
    disk_name: str
    part_num: str
    site_name: str


class MockDisk(NamedTuple):
    children: Iterable


class PostRequestModel(NamedTuple):
    expect: List
    request: Optional[dict] = None
    lock: bool = False


class UmountDockerPathModel(NamedTuple):
    expect: bool
    output: List[Tuple[int, str]]
    cmd: int = 0


class MountDockerPath(NamedTuple):
    expect: bool
    cmd: int


class MountDiskModel(NamedTuple):
    expect: Optional[Type[DevError]]
    mount_point: Optional[str] = None
    mount_path: str = ""
    white: bool = False
    not_sub: bool = False
    exists: bool = False
    is_root: bool = True
    cmd: Tuple[int, str] = (0, "")


class UmountDiskModel(NamedTuple):
    expect: Optional[Type[DevError]]
    mount_point: str
    white: bool = True
    umount_docker: bool = True
    is_root: bool = True
    cmd: int = 0


class PatchRequestModel(NamedTuple):
    expect: List
    request: dict = {}
    part_name: str = ""
    lock: bool = False
    primary: bool = False
    part: Any = None
    mount: Any = None
    umount: Any = None


class TestPartition:
    use_cases = {
        "test_partition_id_check": {
            "valid_name": PartIdCheckModel(expect=True, part_id="sda"),
            "null_name": PartIdCheckModel(expect=False, part_id=None),
            "empty_name": PartIdCheckModel(expect=False, part_id=""),
            "too_long": PartIdCheckModel(expect=False, part_id="s" * 129),
        },
        "test_post_request": {
            "busy": PostRequestModel(expect=[CommonMethods.ERROR, ERROR_MESSAGE.get("busy")], lock=True),
            "null_request": PostRequestModel(expect=[CommonMethods.ERROR, ERROR_MESSAGE.get("param_wrong")]),
            "500_emmc": PostRequestModel([CommonMethods.ERROR, ERROR_MESSAGE.get("param_wrong")],
                                         request={
                                             "Number": 1,
                                             "CapacityBytes": "10",
                                             "Links": [{"Device": {"@odata.id": "/dev/mmcblk0"}}],
                                             "FileSystem": "ext4"
                                         }),
            "space_not_enough": PostRequestModel(expect=[CommonMethods.ERROR, ERROR_MESSAGE.get("not_enough_space")],
                                                 request={
                                                     "Number": 1,
                                                     "CapacityBytes": "10000000000",
                                                     "Links": [{"Device": {"@odata.id": "/dev/sda"}}],
                                                     "FileSystem": "ext4"
                                                 }),
            "partition_num_error": PostRequestModel(
                expect=[CommonMethods.ERROR, ERROR_MESSAGE.get("partition_num_error")],
                request={
                    "Number": 15,
                    "CapacityBytes": "10",
                    "Links": [{"Device": {"@odata.id": "/dev/sda"}}],
                    "FileSystem": "ext4"
                }),
            "ok": PostRequestModel(
                expect=[CommonMethods.OK, "OK"],
                request={
                    "Number": 1,
                    "CapacityBytes": "10",
                    "Links": [{"Device": {"@odata.id": "/dev/sda"}}],
                    "FileSystem": "ext4"
                })
        },
        "test_umount_docker_path": {
            "get container failed": UmountDockerPathModel(expect=False, output=[(1, ""), ]),
            "containers exists": UmountDockerPathModel(expect=False, output=[(0, "abc")]),
            "stop docker failed": UmountDockerPathModel(expect=False, output=[(0, "")], cmd=1),
            "umount failed": UmountDockerPathModel(expect=False, output=[(0, ""), (1, "")], cmd=0),
            "umount success": UmountDockerPathModel(expect=True, output=[(0, ""), (0, "")], cmd=0),
        },
        "test_mount_docker_path": {
            "start docker failed": MountDockerPath(expect=False, cmd=1),
            "start docker success": MountDockerPath(expect=True, cmd=0),
        },
        "test_mount_disk": {
            "mounted": MountDiskModel(expect=PartMounted, mount_point="/dev/abc"),
            "not white": MountDiskModel(expect=PathNotInWhite, white=False),
            "sub of mounted": MountDiskModel(expect=MountPathInvalid, white=True, mount_path="/opt/mount",
                                             not_sub=False),
            "path_exists": MountDiskModel(expect=MountPathExisted, white=True, mount_path="/opt/mount", not_sub=True,
                                          exists=True),
            "mount_failed": MountDiskModel(expect=MountFailed, white=True, mount_path="/opt/mount", not_sub=True,
                                           exists=False, cmd=(1, "")),
            "mount_success": MountDiskModel(expect=None, white=True, mount_path="/opt/mount", not_sub=True,
                                            exists=False, cmd=(0, "")),
        },
        "test_unmount_disk": {
            "not mounted": UmountDiskModel(expect=PartNotMount, mount_point=""),
            "not white": UmountDiskModel(expect=PathNotInWhite, mount_point="/opt/mount", white=False),
            "umount docker failed": UmountDiskModel(expect=UmountDockerFailed, mount_point="/var/lib/docker",
                                                    white=True, umount_docker=False),
            "umount failed": UmountDiskModel(expect=UmountFailed, mount_point="/opt/mount", white=True, cmd=1),
            "umount success": UmountDiskModel(expect=None, mount_point="/opt/mount", white=True, cmd=0)
        },
        "test_patch_request": {
            "invalid part name": PatchRequestModel(
                expect=[CommonMethods.ERROR, ERROR_MESSAGE.get("partition_name_wrong")]),
            "locked": PatchRequestModel(expect=[CommonMethods.ERROR, ERROR_MESSAGE.get("busy")], lock=True,
                                        part_name="sda"),
            "operator_illegal": PatchRequestModel(expect=[CommonMethods.ERROR, ERROR_MESSAGE.get("operator_illegal")],
                                                  lock=False, part_name="sda", request={"_User": "***", "_Xip": "__"}),
            "load part failed": PatchRequestModel(expect=[CommonMethods.NOT_EXIST, "sda not exists."], part_name="sda",
                                                  lock=False, part=[Exception]),
            "mount failed": PatchRequestModel(expect=[CommonMethods.ERROR, [110610, 'Mount partition failed.']],
                                              part_name="sda", lock=False, part=[None],
                                              request={"MountPath": "/dev/sda"}, mount=[Exception]),
            "primary": PatchRequestModel(expect=[CommonMethods.NOT_EXIST, "input para invalid"],
                                         part_name="sda", lock=False, part=[None], primary=True),
            "mount success": PatchRequestModel(expect=[CommonMethods.OK, "Mounted successfully"],
                                               part_name="sda", lock=False, part=[None],
                                               request={"MountPath": "/dev/sda"}, mount=[None]),
            "umount failed": PatchRequestModel(expect=[CommonMethods.ERROR, [110604, 'Unmount partition failed.']],
                                               part_name="sda", lock=False, part=[None], umount=[Exception]),
            "umount success": PatchRequestModel(expect=[CommonMethods.OK, "Umount successfully"], part_name="sda",
                                                lock=False, part=[None], umount=[None])
        }
    }

    @staticmethod
    def test_partition_id_check(model: PartIdCheckModel):
        assert bool(Partition._partition_id_check(model.part_id)) == model.expect

    @staticmethod
    def test_get_part_detail(mocker: MockerFixture):
        cur_dir = Path(__file__).parent
        lsblk_info = json.loads(cur_dir.joinpath("lsblk.json").read_text())["blockdevices"][0]["children"][0]
        udev_sda1 = cur_dir.joinpath("sda1_udev_info.txt").read_text()
        mocker.patch.object(DeviceLoader, "_load_block_info", return_value=lsblk_info)
        mocker.patch.object(DeviceLoader, "exec_cmd", return_value=(0, udev_sda1))
        mocker.patch("lib.Linux.systems.disk.device_loader.SystemUtils").return_value.is_a500 = True
        mocker.patch.object(Partition, "_part_is_primary", return_value=False)
        mocker.patch.object(Partition, "_refresh_cache")
        part = Partition()
        part.part_cache.append("sda1")
        part.get_all_info("sda1")
        assert (part.Name, str(part.CapacityBytes), part.DeviceName) == ("sda1", "1073742336", "/dev/sda")
        part.part_cache.clear()

    @staticmethod
    def test_post_request(mocker: MockerFixture, model: PostRequestModel):
        mocker.patch.object(Partition, "lock").locked.return_value = model.lock
        cur_dir = Path(__file__).parent
        lsblk_info = cur_dir.joinpath("lsblk.json").read_text()
        udev_sda = cur_dir.joinpath("sda_udev_info.txt").read_text()
        udev_sda1 = cur_dir.joinpath("sda1_udev_info.txt").read_text()
        udev_sda2 = cur_dir.joinpath("sda2_udev_info.txt").read_text()
        effect = ((0, lsblk_info), (0, udev_sda), (0, udev_sda1), (0, udev_sda2))
        mocker.patch.object(DeviceLoader, "exec_cmd", side_effect=effect)
        mocker.patch.object(Partition, "_create_partitions")
        mocker.patch.object(Partition, "_refresh_cache")
        part = Partition()
        part.disk_cache = ["sda", "mmcblk0"]
        assert part.post_request(model.request) == model.expect
        part.disk_cache.clear()

    @staticmethod
    def test_delete_request(mocker: MockerFixture):
        mocker.patch.object(Partition, "lock").locked.return_value = False
        cur_dir = Path(__file__).parent
        lsblk_info = json.loads(cur_dir.joinpath("lsblk.json").read_text())["blockdevices"][0]["children"][0]
        udev_sda1 = cur_dir.joinpath("sda1_udev_info.txt").read_text()
        mocker.patch.object(DeviceLoader, "_load_block_info", return_value=lsblk_info)
        mocker.patch.object(DeviceLoader, "exec_cmd", return_value=(0, udev_sda1))
        mocker.patch.object(Partition, "_part_is_primary", return_value=False)
        mocker.patch.object(Partition, "_rm_part")
        mocker.patch.object(Partition, "_refresh_cache")
        part = Partition()
        part.part_cache.append("sda1")
        assert part.delete_request({}, "sda1") == [CommonMethods.OK, "OK"]

    @staticmethod
    def test_umount_docker_path(mocker: MockerFixture, model: UmountDockerPathModel):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output", side_effect=model.output)
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=model.cmd)
        assert Partition.umount_docker_path("") == model.expect

    @staticmethod
    def test_mount_docker_path(mocker: MockerFixture, model: MountDockerPath):
        mocker.patch("lib.Linux.systems.disk.partition.MefInfo")
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=model.cmd)
        assert Partition.mount_docker_path() == model.expect

    @staticmethod
    def test_mount_disk(mocker: MockerFixture, model: MountDiskModel):
        part = mocker.MagicMock(mount_point=model.mount_point)
        mocker.patch.object(FileCheck, "check_path_is_root", return_value=model.is_root)
        mocker.patch.object(Partition, "check_path_whitelist", return_value=model.white)
        mocker.patch.object(Partition, "check_mount_path_is_subdirectory_of_mounted_path",
                            return_value=model.not_sub)
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch("os.makedirs")
        mocker.patch.object(ExecCmd, "exec_cmd_get_output", return_value=model.cmd)
        mocker.patch.object(Partition, "remove_mount_path")
        mocker.patch.object(Partition, "_refresh_cache")
        if model.expect:
            with pytest.raises(model.expect):
                Partition().mount_disk(model.mount_path, part)
        else:
            Partition().mount_disk(model.mount_path, part)

    @staticmethod
    def test_unmount_disk(mocker: MockerFixture, model: UmountDiskModel):
        part = mocker.MagicMock(mount_point=model.mount_point, path="")
        mocker.patch.object(FileCheck, "check_path_is_root", return_value=model.is_root)
        mocker.patch.object(Partition, "check_path_whitelist", return_value=model.white)
        mocker.patch.object(Partition, "umount_docker_path", return_value=model.umount_docker)
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=model.cmd)
        mocker.patch.object(Partition, "_refresh_cache")
        if model.expect:
            with pytest.raises(model.expect):
                Partition().unmount_disk(part)
        else:
            Partition().unmount_disk(part)

    @staticmethod
    def test_patch_request(mocker: MockerFixture, model: PatchRequestModel):
        mocker.patch.object(Partition, "lock").locked.return_value = model.lock
        mocker.patch.object(DeviceLoader, "load_part", side_effect=model.part)
        mocker.patch.object(Partition, "mount_disk", side_effect=model.mount)
        mocker.patch.object(Partition, "_part_is_primary", return_value=model.primary)
        mocker.patch.object(Partition, "unmount_disk", side_effect=model.umount)
        mocker.patch.object(Partition, "_refresh_cache")
        part = Partition()
        part.part_cache.append("sda")
        assert part.patch_request(model.request, model.part_name) == model.expect
