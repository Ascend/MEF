from collections import namedtuple

from common.checkers import CheckResult
from pytest_mock import MockerFixture
from lib.Linux.systems.logger_collect import LoggerCollect
from common.common_methods import CommonMethods

ExecLogCollectShell = namedtuple("ExecLogCollectShell", "expect, free, out")
CheckExternalParameter = namedtuple("CheckExternalParameter", "expect, check")
PostRequest = namedtuple("PostRequest", "expect, check, request, cmd, sign, exists, path_valid, group")


class TestLoggerCollect:
    use_cases = {
        "test_check_external_parameter": {
            "exception": (
                [CommonMethods.ERROR, "Check external parameter failed. Because of Parameter is invalid"], Exception
            ),
            "invalid params": (
                [CommonMethods.ERROR, "checker_class failed. Because of Parameter is invalid"], [CheckResult(False, "")]
            ),
            "OK": (
                [CommonMethods.OK, ], [CheckResult(True, ""), ]
            ),
        },
        "test_post_request": {
            "invalid params": (
                [CommonMethods.ERROR, "Check input para failed, [400, 'invalid params']"],
                [CommonMethods.ERROR, "invalid params"], None, None, None, None, None, None
            ),
            "Collect log success.": (
                [CommonMethods.OK, "Collect log success."], [CommonMethods.OK, ], {}, "", [0], True, True, [True]
            )
        }
    }

    @staticmethod
    def test_check_external_parameter(mocker: MockerFixture, model: CheckExternalParameter):
        checker_class = mocker.MagicMock()
        checker_class.return_value.check.side_effect = model.check
        assert LoggerCollect.check_external_parameter(checker_class, "") == model.expect

    @staticmethod
    def test_post_request(mocker: MockerFixture, model: PostRequest):
        mocker.patch.object(LoggerCollect, "check_external_parameter", return_value=model.check)
        mocker.patch.object(LoggerCollect, "collect_log")
        assert LoggerCollect().post_request(model.request) == model.expect
