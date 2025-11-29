# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from collections import namedtuple
from configparser import ConfigParser
from pathlib import Path

from pytest_mock import MockerFixture

from bin.environ import Env
from common.file_utils import FileCheck
from common.utils.result_base import Result
from prepare_docker_root_dir import DockerRootDirPrepare


class TestUtils:
    name = ""

    @staticmethod
    def as_posix():
        return False

    @staticmethod
    def exists():
        return False

    @staticmethod
    def is_block_device():
        return False


class TestUtilsTrue:
    @staticmethod
    def exists():
        return True

    @staticmethod
    def is_block_device():
        return True


class TestPrepareDockerRootDir:
    ExecuteCase = namedtuple("ExecuteCase", "expected, valid")
    PersistCase = namedtuple("PersistCase", "check_input_path_valid")
    MountPointInPersistCase = namedtuple("MountPointInPersistCase", "expected, check_path_is_exist_and_valid")
    ValidCase = namedtuple("ValidCase", "expected, mount_point_in_persist, get_partition_by_label")

    use_cases = {
        "test_execute": {
            "valid_is_false": (1, False),
            "test_execute": (0, True),
        },
        "test_persist": {
            "normal": (Result(True), ),
            "return": (Result(False), ),
        },
        "test_mount_point_in_persist": {
            "normal": (False, Result(True),),
            "first_return": (False, Result(False),),
        },
        "test_valid": {
            "first_return": (False, Result(True), TestUtils),
            "second_return": (False, Result(False), TestUtils),
            "third_return": (False, Result(False), TestUtilsTrue),
        },
    }

    def test_execute(self, mocker: MockerFixture, model: ExecuteCase):
        mocker.patch.object(DockerRootDirPrepare, "valid", return_value=model.valid)
        mocker.patch.object(DockerRootDirPrepare, "persist", return_value=None)
        ret = DockerRootDirPrepare.execute(DockerRootDirPrepare())
        assert model.expected == ret

    def test_persist(self, mocker: MockerFixture, model: PersistCase):
        mocker.patch.object(FileCheck, "check_input_path_valid", return_value=model.check_input_path_valid)
        mocker.patch.object(Env, "get_partition_by_label", return_value=TestUtils)
        mocker.patch.object(ConfigParser, "add_section", return_value=None)
        mocker.patch.object(ConfigParser, "set", return_value=None)
        mocker.patch("os.fdopen").return_value.__enter__.return_value.write.side_effect = "abc"
        mocker.patch.object(ConfigParser, "write", return_value=True)
        ret = DockerRootDirPrepare.persist(DockerRootDirPrepare())
        assert not ret

    def test_mount_point_in_persist(self, mocker: MockerFixture, model: MountPointInPersistCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid",
                            return_value=model.check_path_is_exist_and_valid)
        mocker.patch.object(ConfigParser, "read", return_value=True)
        mocker.patch.object(ConfigParser, "sections", return_value=[0])
        mocker.patch.object(ConfigParser, "options", return_value=[0])
        mocker.patch.object(ConfigParser, "get", return_value=1)
        mocker.patch.object(ConfigParser, "has_section", return_value=False)
        ret = DockerRootDirPrepare.mount_point_or_mount_part_in_persist(DockerRootDirPrepare(), Path("/dev"))
        assert model.expected == ret

    def test_valid(self, mocker: MockerFixture, model: ValidCase):
        mocker.patch.object(
            DockerRootDirPrepare, "mount_point_or_mount_part_in_persist", return_value=model.mount_point_in_persist
        )
        mocker.patch.object(Env, "get_partition_by_label", return_value=model.get_partition_by_label)
        mocker.patch.object(DockerRootDirPrepare, "MOUNT_POINT", return_value=model.get_partition_by_label)
        ret = DockerRootDirPrepare.valid(DockerRootDirPrepare())
        assert model.expected == ret
