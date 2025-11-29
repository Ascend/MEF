
from common.checkers.fd_param_checker import HostnameChecker
from fd_msg_process.midware_urls import MidwareUris


class TestMidwareurlsConfigHostname:
    def test_hostname_external_param_should_ok_when_hostname_check_ok(self):
        payload_publish = {
            "topic": "config_hostname",
            "percentage": "0%",
            "result": "failed",
            "reason": ""
        }
        ret = MidwareUris().check_external_param(
            HostnameChecker, {"hostname": "abcd123"}, payload_publish, "Invalid hostname. %s")
        assert ret[0] == 0


