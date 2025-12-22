from collections import namedtuple

from pytest_mock import MockerFixture

from common.utils.app_common_method import AppCommonMethod
from upgrade_service.models import UpgradeInfo
from upgrade_service.upgrade_report_manager import UpgradeReportManager


class TestUpgradeService:
    use_cases = {
        "test_report_error_message_to_mqtt": {
            "normal": ("ERR.0100, test", 100, "test", "0100"),
        },
        "test_report_upgrade_process": {
            "normal": ("100%", UpgradeInfo("processing", 100, "NA", ""), 1)
        },
        "test_update_error_message": {
            "normal": ("failed", 100, "test"),
        }
    }

    UpgradeReportManagerCase2 = namedtuple("UpgradeReportManagerCase", "expect, input_para1, input_para2")
    UpgradeReportManagerCase3 = namedtuple("UpgradeReportManagerCase", "expect, input_para1, input_para2, input_para3")

    @staticmethod
    def test_report_error_message_to_mqtt(mocker: MockerFixture, model: UpgradeReportManagerCase3):
        instance = UpgradeReportManager()
        mocker.patch.object(AppCommonMethod, "convert_err_code_fd_format", return_value=model.input_para3)
        instance.report_failed_message(model.input_para1, model.input_para2)
        assert model.expect == instance.report_model.reason

    @staticmethod
    def test_report_upgrade_process(model: UpgradeReportManagerCase2):
        instance = UpgradeReportManager()
        instance.report_upgrade_process(model.input_para1)
        assert model.expect == instance.report_model.percentage

    @staticmethod
    def test_update_error_message(model: UpgradeReportManagerCase2):
        instance = UpgradeReportManager()
        instance.report_failed_message(model.input_para1, model.input_para2)
        assert model.expect == instance.report_model.result

