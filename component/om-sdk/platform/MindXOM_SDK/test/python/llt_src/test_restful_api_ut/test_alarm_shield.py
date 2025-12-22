from unittest.mock import patch, MagicMock
import mock

from common.file_utils import FileCheck
from lib.Linux.systems.Alarm.alarm_shield import AlarmShield


class TestKeyUpdater:
    FAILED = 400
    OK = 200

    def test_route(self):
        ret = AlarmShield().patch_request("123")
        assert ret[0] == self.FAILED

    def test_check_alarm_shield_should_failed_when_param_is_int(self):
        ret = AlarmShield.check_alarm_shield(2)
        assert ret[0] == self.FAILED

    def test_check_alarm_shield_should_failed_when_param_is_not_dict(self):
        ret = AlarmShield.check_alarm_shield([1, 2])
        assert ret[0] == self.FAILED and ret[
            1] == "Alarm masking rules are illegal,alarm_clear_messages_dict is not dict"

    def test_check_alarm_shield_should_failed_when_dict_key_is_illegal(self):
        ret = AlarmShield.check_alarm_shield([{"UniquelyIdentifies": "a000000000",
                                                "Ala": "00000000",
                                                "PerceivedSeverity": "2",
                                                "AlarmInstance": "M.2"}])
        assert ret[0] == self.FAILED and ret[1] == "Alarm masking key are illegal"

    @patch.object(AlarmShield, 'get_all_alarm_role', mock.Mock(return_value=[
        {"UniquelyIdentifies": "a000000000",
         "AlarmId": "00000000",
         "PerceivedSeverity": "2",
         "AlarmInstance": "M.2"},
        {"UniquelyIdentifies": "a000000001",
         "AlarmId": "00000001",
         "PerceivedSeverity": "2",
         "AlarmInstance": "M.2"}
    ]))
    def test_check_alarm_shield_should_failed_when_rules_are_illegal(self):
        ret = AlarmShield.check_alarm_shield([{"UniquelyIdentifies": "a000000000",
                                                "AlarmId": "000000320",
                                                "PerceivedSeverity": "2",
                                                "AlarmInstance": "M.2"}])
        assert ret[0] == self.FAILED and ret[1] == "Alarm masking rules are illegal,not in rules"

    @patch.object(AlarmShield, 'get_all_alarm_role', mock.Mock(return_value=[
        {"UniquelyIdentifies": "a000000000",
         "AlarmId": "00000000",
         "PerceivedSeverity": "2",
         "AlarmInstance": "M.2"},
        {"UniquelyIdentifies": "a000000001",
         "AlarmId": "00000001",
         "PerceivedSeverity": "2",
         "AlarmInstance": "M.2"}
    ]))
    def test_check_alarm_shield_should_ok_when_check_is_ok(self):
        ret = AlarmShield.check_alarm_shield([
            {"UniquelyIdentifies": "a000000000",
             "AlarmId": "00000000",
             "PerceivedSeverity": "2",
             "AlarmInstance": "M.2"}])
        assert ret[0] == self.OK

    @patch.object(AlarmShield, 'check_alarm_shield', mock.Mock(return_value=[400, ]))
    def test_patch_request_should_failed_when_check_failed(self):
        ret = AlarmShield().patch_request({"UniquelyIdentifies": "a000000000",
                                           "Ala": "00000000",
                                           "PerceivedSeverity": "2",
                                           "AlarmInstance": "M.2"})
        assert ret[0] == self.FAILED

    @patch.object(AlarmShield, 'write_alarm_shield', mock.Mock(return_value=200))
    @patch.object(AlarmShield, 'check_alarm_shield', mock.Mock(return_value=[200, ]))
    def test_patch_request_invalid(self):
        ret = AlarmShield().patch_request({"UniquelyIdentifies": "a000000000",
                                           "Ala": "00000000",
                                           "PerceivedSeverity": "2",
                                           "AlarmInstance": "M.2"})
        assert ret == [500, 'Alarm config modify failed because of invalid request.']

    def test_write_alarm_shield_should_failed_param_is_none(self):
        ret = AlarmShield().write_alarm_shield(None)
        assert ret[0] == self.FAILED

    @patch("os.fsync", MagicMock())
    @patch("os.open", MagicMock())
    @patch("os.fdopen", MagicMock())
    @patch.object(FileCheck, 'check_is_link_exception', mock.Mock(return_value=True))
    def test_write_alarm_shield_should_ok_param_is_ok(self):
        ret = AlarmShield().write_alarm_shield([
            {"UniquelyIdentifies": "a000000000",
             "AlarmId": "00000000",
             "PerceivedSeverity": "2",
             "AlarmInstance": "M.2"}])
        assert ret[0] == self.OK
