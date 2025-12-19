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
from collections import namedtuple
from pathlib import Path
from unittest.mock import patch

import pytest
from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.utils.result_base import Result
from mount_white_path import WhitelistError, is_sub_dir, add_whitelist_path, delete_whitelist_path, \
    display_whitelist_path, path_in_whitelist

ReadTimeZoneCfgCase = namedtuple("ReadTimeZoneCfgCase", "expected, sub, version_info")
AddWhitelistPathCase = namedtuple("AddWhitelistPathCase", "")
DeleteWhitelistPathCase = namedtuple("DeleteWhitelistPathCase", "expected, normpath, is_mount")
PathInWhitelistCase = namedtuple("PathInWhitelistCase", "path_in_whitelist")


class TestUtils:
    path = "/test"


class TestMountWhitePath:
    use_cases = {
        "test_read_time_zone_cfg": {
            "normal": (True, "/home/user/Documents/example.txt", (3, 8)),
            "ValueError": (False, "/var/log", (3, 8)),
        },
        "test_read_time_zone_cfg_exception": {},
        "test_add_whitelist_path": {},
        "test_delete_whitelist_path_raise": {
            "first": (None, "/opt/mount", False),
            "second": (None, "/opt", True),
            "third": (None, "/opt", False),
        },
        "test_delete_whitelist_path": {
           "normal": (None, "/opt", False),
        },
        "test_path_in_whitelist": {
            "normal": (True, ),
            "final": (False,),
        },
    }

    def test_read_time_zone_cfg(self, model: ReadTimeZoneCfgCase):
        with patch("sys.version_info", model.version_info):
            ret = is_sub_dir(model.sub, "/home/user")
            assert model.expected == ret

    def test_add_whitelist_path(self, mocker: MockerFixture):
        mocker.patch("lib.Linux.systems.disk.mount_mgr.get_whitelist_path_count", return_value=63)
        mocker.patch("lib.Linux.systems.disk.mount_mgr.path_in_whitelist", return_value=False)
        mocker.patch("lib.Linux.systems.disk.mount_mgr.add_whitelist_path", return_value=None)
        mocker.patch("lib.Linux.systems.disk.mount_mgr.query_whitelist_path", return_value=[TestUtils, ])
        mocker.patch("mount_white_path.is_sub_dir", return_value=False)
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=Result(result=True))
        assert not add_whitelist_path("/home/test/data")

    def test_delete_whitelist_path_raise(self, mocker: MockerFixture, model: DeleteWhitelistPathCase):
        mocker.patch("os.path.normpath", return_value=model.normpath)
        mocker.patch.object(Path, "is_mount", return_value=model.is_mount)
        mocker.patch("lib.Linux.systems.disk.mount_mgr.delete_mount_white_path", return_value=False)
        with pytest.raises(WhitelistError):
            delete_whitelist_path("/")

    def test_delete_whitelist_path(self, mocker: MockerFixture, model: DeleteWhitelistPathCase):
        mocker.patch("os.path.normpath", return_value=model.normpath)
        mocker.patch.object(Path, "is_mount", return_value=model.is_mount)
        mocker.patch("lib.Linux.systems.disk.mount_mgr.delete_mount_white_path", return_value=True)
        assert not delete_whitelist_path("/")

    def test_display_whitelist_path(self, mocker: MockerFixture):
        mocker.patch("lib.Linux.systems.disk.mount_mgr.query_whitelist_path", return_value=[TestUtils, ])
        assert not display_whitelist_path()

    def test_path_in_whitelist(self, mocker: MockerFixture, model: PathInWhitelistCase):
        mocker.patch("lib.Linux.systems.disk.mount_mgr.path_in_whitelist", return_value=model.path_in_whitelist)
        assert not path_in_whitelist("/home/test/data")
