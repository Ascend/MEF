from collections import namedtuple

from common.utils.system_utils import SystemUtils

from bin.monitor_config import SystemSetting
from pytest_mock import MockerFixture
from ut_utils.mock_utils import mock_cdll


with mock_cdll():
    from lib.Linux.systems.ai_processor import AiProcessorInfo
    from devm.device_mgr import DEVM
    from devm.exception import DeviceManagerError

AiProcessorInfoCase = namedtuple("AiProcessorInfoCase", "expect, path_valid, read_data")
CheckGetAllEdgeSystemInfo = namedtuple("CheckGetAllEdgeSystemInfo", "expect, get_putilization_rate npu_version")
CheckGetMemorySize = namedtuple("CheckGetMemorySize", "expect, get_mem_size")
CheckGetVendor = namedtuple("CheckGetVendor", "expect, get_vendor_info")
CheckGetModel = namedtuple("CheckGetModel", "expect, get_model_info")
CheckGetAbilityDesc = namedtuple("CheckGetAbilityDesc", "expect, get_ability_desc_info")
CheckGetChipInfo = namedtuple("CheckGetChipInfo", "expect")
CheckGetAiCount = namedtuple("CheckGetAiCount", "expect, npu")
CheckGetMinidHealthStatus = namedtuple("CheckGetMinidHealthStatus", "expect, get_sys_health")
CheckGetPutilizationRate = namedtuple("CheckGetPutilizationRate", "expect, get_value")


class TestAiProcessor:
    use_cases = {
        "test_get_ai_count": {
            "normal": (1, [1]),
            "exception": (None, DeviceManagerError()),
        },
        "test_get_chip_info": {
            "normal": (None, ),

        },
        "test_get_minid_health_status": {
            "invalid_when_res_is_0": ("OK", [0]),
            "invalid_when_res_is_1": ("General warning", [1]),
            "invalid_when_res_is_2": ("Important warning", [2]),
            "invalid_when_res_is_3": ("Urgent warning", [3]),
            "invalid_when_res_is_minustwo": (None, [-2]),
            "others": ("Unknown mistake", [-100]),
            "exception": (None, DeviceManagerError())
        },
        "test_get_model": {
            "normal": ("64", [{"model": "64"}]),
            "exception": (None, DeviceManagerError()),
        },
        "test_get_vendor": {
            "normal": ("Huawei", [{"vendor": "Huawei"}]),
            "exception": (None, DeviceManagerError())
        },
        "test_get_all_info": {
            "normal": (1, [1, 1, 1, 1, 1], "npu-23.0"),
        },
        "test_get_memory_size": {
            "normal": (1.0, [1000]),
            "exception": (None, DeviceManagerError()),
        },
        "test_get_putilization_rate": {
            "normal": (2, {"test": 2}),
        }
    }

    @staticmethod
    def test_get_putilization_rate(mocker: MockerFixture, model: CheckGetPutilizationRate):
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.return_value = model.get_value
        ai_processor_info = AiProcessorInfo()
        ret = ai_processor_info.get_usage_rate("test")
        assert ret == model.expect

    @staticmethod
    def test_get_minid_health_status(mocker: MockerFixture, model: CheckGetMinidHealthStatus):
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.side_effect = model.get_sys_health
        ai_processor_info = AiProcessorInfo()
        ai_processor_info.get_minid_health_status()
        assert ai_processor_info.Health == model.expect

    @staticmethod
    def test_get_ai_count(mocker: MockerFixture, model: CheckGetAiCount):
        mocker.patch.object(DEVM, "count_devices_by_module", side_effect=model.npu)
        ai_processor_info = AiProcessorInfo()
        ai_processor_info.get_ai_count()
        assert ai_processor_info.Count == model.expect

    @staticmethod
    def test_get_chip_info(mocker: MockerFixture, model: CheckGetChipInfo):
        mocker.patch.object(AiProcessorInfo, "get_model")
        mocker.patch.object(AiProcessorInfo, "get_vendor")
        mocker.patch.object(AiProcessorInfo, "get_ability_desc")
        ai_processor_info = AiProcessorInfo()
        assert model.expect == ai_processor_info.get_chip_info()

    @staticmethod
    def test_get_all_info(mocker: MockerFixture, model: CheckGetAllEdgeSystemInfo):
        mocker.patch.object(AiProcessorInfo, "get_chip_info")
        mocker.patch.object(AiProcessorInfo, "get_ai_count")
        mocker.patch.object(AiProcessorInfo, "get_memory_size")
        mocker.patch.object(AiProcessorInfo, "get_minid_health_status")
        mocker.patch.object(AiProcessorInfo, "get_usage_rate", side_effect=model.get_putilization_rate)
        mocker.patch.object(SystemSetting, "get_npu_version", side_effect=model.npu_version)
        ai_processor_info = AiProcessorInfo()
        ai_processor_info.get_all_info()
        assert model.expect == ai_processor_info.DdrBandWidth

    @staticmethod
    def test_get_memory_size(mocker: MockerFixture, model: CheckGetMemorySize):
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.side_effect = model.get_mem_size
        ai_processor_info = AiProcessorInfo()
        ai_processor_info.get_memory_size()
        assert model.expect == ai_processor_info.Ddr

    @staticmethod
    def test_get_vendor(mocker: MockerFixture, model: CheckGetVendor):
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.side_effect = model.get_vendor_info
        ai_processor_info = AiProcessorInfo()
        ai_processor_info.get_vendor()
        assert model.expect == ai_processor_info.Manufacturer

    @staticmethod
    def test_get_model(mocker: MockerFixture, model: CheckGetModel):
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.side_effect = model.get_model_info
        ai_processor_info = AiProcessorInfo()
        ai_processor_info.get_model()
        assert model.expect == ai_processor_info.Model
