from collections import namedtuple

from pytest_mock import MockerFixture

from common.constants.upgrade_constants import UpgradeState
from common.utils.exec_cmd import ExecCmd

from common.common_methods import CommonMethods
from ut_utils.mock_utils import mock_cdll, mock_npu_smi_board_type

with mock_cdll():
    with mock_npu_smi_board_type():
        from lib.Linux.upgrade.upgrade_effect import UpgradeEffect
        from lib.Linux.upgrade.upgrade_new import Upgrade
PostRequest = namedtuple("PostRequest", "expect, lock, request, allow_effect")


class TestUpgradeEffect:
    use_cases = {
        "test_post_request": {
            "busy": ([CommonMethods.ERROR, "ERR.01-Upgrade effect is busy."], True,
                     {"ResetType": "GracefulRestart"}, True),
            "null request": ([CommonMethods.ERROR, "ERR.02-Request data is invalid."], False, {}, True),
            "invalid request": ([CommonMethods.ERROR, "ERR.02-Request data is invalid."], False, {"a": 1}, True),
            "invalid ResetType": ([CommonMethods.ERROR, "ERR.02-Request data is invalid."], False, {"ResetType": None},
                                  True),
            "not allow effect": ([CommonMethods.ERROR, "ERR.04-Upgrade failed, restart is not allowed."], False,
                                 {"ResetType": "GracefulRestart"}, False),
            "normal": ([CommonMethods.OK, "ERR.00-Effect firmware success."], False, {"ResetType": "GracefulRestart"},
                       True),
        }
    }

    @staticmethod
    def test_post_request(mocker: MockerFixture, model: PostRequest):
        mocker.patch.object(UpgradeEffect, "effect_lock").locked.return_value = model.lock
        mocker.patch.object(Upgrade, "allow_effect", return_value=model.allow_effect)
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=0)
        mocker.patch("threading.Thread")
        Upgrade.upgrade_state = UpgradeState.UPGRADE_NO_STATE
        assert UpgradeEffect().post_request(model.request) == model.expect
