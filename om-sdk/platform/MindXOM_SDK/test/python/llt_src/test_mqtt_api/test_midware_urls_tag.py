
from common.checkers.fd_param_checker import SysAssetTagChecker
from fd_msg_process.common_redfish import CommonRedfish
from fd_msg_process.midware_urls import MidwareUris
from common.common_methods import CommonMethods


class TestMidwareurlsTag:

    def test_check_tag_external_param_should_failed_when_tag_check_failed(self):
        payload_publish = {
            "topic": "tag",
            "percentage": "0%",
            "result": "failed",
            "reason": ""
        }
        ret = MidwareUris().check_external_param(
            SysAssetTagChecker, {"asset_tag": "abcd123A[]。"}, payload_publish, "Set system tag failed. %s")
        assert ret[0] == -1

    def test_check_tag_external_param_should_ok_when_tag_check_ok(self):
        payload_publish = {
            "topic": "tag",
            "percentage": "0%",
            "result": "failed",
            "reason": ""
        }
        ret = MidwareUris().check_external_param(
            SysAssetTagChecker, {"asset_tag": "abcd123A[]"}, payload_publish, "Set system tag failed. %s")
        assert ret[0] == 0

    def test_midware_urls_check_status_should_ok_when_status_is_ok(self):
        ret_dict = {}
        ret_dict['status'] = CommonMethods.OK
        ret = CommonRedfish.check_status_is_ok(ret_dict)
        assert ret
