# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from collections import namedtuple

import pytest
from pytest_mock import MockerFixture

from common.common_methods import CommonMethods
from common.exception.biz_exception import BizException
from common.utils.exec_cmd import ExecCmd
from conftest import TestBase
from test_mqtt_api.get_log_info import GetLogInfo
from ut_utils.mock_utils import mock_npu_smi_board_type

with mock_npu_smi_board_type():
    from lib.Linux.upgrade.upgrade_effect import UpgradeEffect
    from lib.Linux.upgrade.upgrade_new import Upgrade
    from lib.Linux.systems.actions import SystemAction, RestoreDefaultsAction

getLog = GetLogInfo()


class TestActions(TestBase):
    use_cases = {
        "test_post_request": {
            "locked": ([CommonMethods.ERROR, "ERR.01-Restart system failed"], None, True, None, None),
            "invalid action": ([CommonMethods.ERROR, "ERR.01-Restart system failed"],
                               {"ResetType": "test"}, False, None, None),
            "exception": ([CommonMethods.ERROR, "ERR.01-Restart system failed"],
                          {"ResetType": "GracefulRestart"}, False, True, Exception),
            "not effect": ([CommonMethods.OK, "ERR.00-Restart system successfully."],
                           {"ResetType": "GracefulRestart"}, False, False, None),
            "effect": ([CommonMethods.OK, "ERR.00-Restart system successfully."],
                       {"ResetType": "GracefulRestart"}, False, True, None),
        },
        "test_graceful_restart": {
            "normal": ("", [None, None]),
            "exception": ("reboot system failed, because unknown error", Exception),
        },
    }

    PostRequestCase = namedtuple("PostRequestCase", "expect, request, lock, allow_effect, effect_firmware")
    GracefulRestartCase = namedtuple("GracefulRestartCase", "expect, exec_cmd")

    @staticmethod
    def test_graceful_restart(mocker: MockerFixture, model: GracefulRestartCase):
        mocker.patch.object(SystemAction, "EDGE_SYSTEM_ACTION_LOCK").locked.return_value = True
        mocker.patch.object(ExecCmd, "exec_cmd", side_effect=model.exec_cmd)
        mocker.patch("os.sync")
        mocker.patch("time.sleep")
        getLog.clear_log()
        SystemAction.graceful_restart()
        assert model.expect in getLog.get_log()

    @staticmethod
    def test_post_request(mocker: MockerFixture, model: PostRequestCase):
        mocker.patch.object(SystemAction, "EDGE_SYSTEM_ACTION_LOCK").locked.return_value = model.lock
        mocker.patch.object(Upgrade, "allow_effect", return_value=model.allow_effect)
        mocker.patch.object(UpgradeEffect, "effect_firmware", side_effect=model.effect_firmware)
        mocker.patch("threading.Thread")
        assert SystemAction.post_request(model.request) == model.expect


class TestRestoreDefaultsAction(TestBase):
    use_cases = {
        "test_check_root_pwd": {
            "not exists": (None, None, True, None),
            "not root_pwd": (None, None, False, None),
            "root_pwd not str": (None, 111, False, None),
            "root_pwd too long": (None, "1111111111111111111111111111", False, None),
            "not ok": (None, "1111111111111", False, False),
        },
        "test_record_restore_log": {
            "blockdev setrw failed": (None, [1, None, None, None], None),
            "mount failed": (None, [0, 1, None, None], None),
            "fdopen failed": (None, [0, 0, None, None], Exception),
            "blockdev setro failed": (None, [0, 0, 1, None], None),
            "umount failed": (None, [0, 0, 0, 1], None),
        },
        "test_restore_defaults": {
            "not contain": (None, "2", {"1": "1"}, None, None),
            "upgrade running": (None, "1", {"1": "1"}, 1, None),
            "restore exception": (None, "1", {"1": "1"}, 0, Exception),
            "restore failed due to upgrading": (None, "1", {"1": "1"}, 0, ["device is upgrading", ""]),
            "restore failed due to other reason": (None, "1", {"1": "1"}, 0, ["123", ""]),
        },
    }

    CheckRootPwdCase = namedtuple("CheckRootPwdCase", "expect, root_pwd, exists, authenticate")
    RecordRestoreLogCase = namedtuple("RecordRestoreLogCase", "expect, exec_cmd, fdopen")
    RestoreDefaultsCase = namedtuple("RestoreDefaultsCase", "expect, ethernet, ip_map, state, subprocess")

    @staticmethod
    def test_restore_defaults(mocker: MockerFixture, model: RestoreDefaultsCase):
        mocker.patch.object(RestoreDefaultsAction, "get_eth_ip_map", return_value=model.ip_map)
        mocker.patch.object(Upgrade, "upgrade_state", return_value=model.state)
        mocker.patch.object(RestoreDefaultsAction, "record_restore_log")
        mocker.patch("subprocess.Popen").return_value.__enter__.return_value.stdin.return_value = "111"
        mocker.patch("subprocess.Popen").\
            return_value.__enter__.return_value.stdout.readlines.return_value = model.subprocess
        with pytest.raises(BizException):
            RestoreDefaultsAction().restore_defaults(model.ethernet, "test", "test")

    @staticmethod
    def test_record_restore_log(mocker: MockerFixture, model: RecordRestoreLogCase):
        mocker.patch.object(ExecCmd, "exec_cmd", side_effect=model.exec_cmd)
        mocker.patch("os.fdopen").return_value.__enter__.return_value.write.side_effect = model.fdopen
        with pytest.raises(BizException):
            RestoreDefaultsAction.record_restore_log("test", "test")

    @staticmethod
    def test_check_root_pwd(mocker: MockerFixture, model: CheckRootPwdCase):
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch("lib.Linux.systems.restore_defaults_action_clib.authenticate", return_value=model.authenticate)
        with pytest.raises(BizException):
            RestoreDefaultsAction.check_root_pwd(model.root_pwd)
