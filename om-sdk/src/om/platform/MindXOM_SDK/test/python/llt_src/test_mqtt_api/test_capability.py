from unittest import mock

from mock.mock import MagicMock

from common.file_utils import FileCheck
from common.utils.result_base import Result
from common.constants.config_constants import ConfigPathConstants
from fd_msg_process.capability import Capability


class TestCapability:

    @mock.patch("os.path.exists", mock.Mock(return_value=False))
    def test_get_capability_from_file_should_failed_when_path_not_exists(self):
        ret = Capability.get_capability_from_file(ConfigPathConstants.DEFAULT_CAPABILITY_FILE)
        assert ret == {}

    @mock.patch.object(FileCheck, 'check_path_is_exist_and_valid',
                       mock.Mock(return_value=Result(result=False, err_msg="path is not exists")))
    @mock.patch("os.path.exists", mock.Mock(return_value=True))
    def test_get_capability_from_file_should_failed_when_check_is_false(self):
        ret = Capability.get_capability_from_file(ConfigPathConstants.DEFAULT_CAPABILITY_FILE)
        assert ret == {}

    @mock.patch("os.path.exists", mock.Mock(return_value=True))
    @mock.patch.object(FileCheck, "check_path_is_exist_and_valid", mock.Mock(return_value=True))
    @mock.patch("builtins.open", mock.Mock(return_value=MagicMock()))
    @mock.patch("json.load", mock.Mock(return_value={'esp_enable': True}))
    def test_get_capability_from_file_should_ok_when_path_is_ok(self):
        ret = Capability.get_capability_from_file(ConfigPathConstants.DEFAULT_CAPABILITY_FILE)
        assert ret == {
            'esp_enable': True
        }

    @mock.patch.object(Capability, 'get_capability_from_file', mock.Mock(return_value={
        "esp_enable": True,
        "product_capability": [
            "profile",
            "Assettag"]}))
    def test_get_default_capability_should_return_dict(self):
        ret = Capability.get_default_capability()
        assert ret == {
            "esp_enable": True,
            "product_capability": [
                "profile",
                "Assettag"]}

    def test_combine_product_capability_should_return_null_when_param_is_none(self):
        ret = Capability.combine_product_capability({})
        assert ret == {}
