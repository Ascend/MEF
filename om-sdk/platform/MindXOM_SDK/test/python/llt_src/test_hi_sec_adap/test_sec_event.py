from collections import namedtuple

import mock
from pytest_mock import MockerFixture

from bin.HiSecAdap.constant import MsgErrorCode
from bin.HiSecAdap.msg import MsgParse
from bin.HiSecAdap.sec_event import HiSecEventProc, MsgSecEventReq, MsgSecEventRsp
from conftest import TestBase
from test_mqtt_api.get_log_info import GetLogInfo
from test_mqtt_api.get_log_info import GetOperationLog

getLog = GetLogInfo()
getOplog = GetOperationLog()


class TestMsgSecEventReq(TestBase):
    use_cases = {
        "test_deal_with": {
            "is_parse_ok_is_false_and_is_check_ok_is_true": (MsgErrorCode.MSG_PARSE_ERROR, False, True),
            "is_parse_ok_is_true_and_is_check_ok_is_false": (MsgErrorCode.MSG_PARA_CHECK_ERROR, True, False),
            "is_parse_ok_is_true_and_is_check_ok_is_true": (MsgErrorCode.MSG_PARSE_OR_PACKER_OK, True, True)
        },
        "test_parse_msg": {
            "normal": ([], []),
        },
        "test_check_msg_req_para": {
            "tag_not_in_sec_eye_event_type": (False, True, [{"tag": 1}]),
            "tag_in_sec_eye_event_type": (True, True, [{"tag": 0x0061}])
        },
        "test_report_to_fd": {
            "check_log_content_valid_is_false": (None, False, None),
            "check_log_content_valid_is_true": (None, True, None)
        },
    }

    MsgSecEventReqCase1 = namedtuple("MsgSecEventReqCase", "expect, input_para")
    MsgSecEventReqCase2 = namedtuple("MsgSecEventReqCase", "expect, input_para1, input_para2")

    def test_deal_with(self, mocker: MockerFixture, model: MsgSecEventReqCase2):
        instance = MsgSecEventReq(b'test')
        instance.is_parse_ok = model.input_para1
        instance._is_check_ok = model.input_para2
        mocker.patch.object(MsgSecEventReq, "_check_msg_req_para", return_value=None)
        mocker.patch.object(MsgSecEventReq, "_report_to_fd", return_value=None)
        assert model.expect == instance.deal_with()

    def test_parse_msg(self, mocker: MockerFixture, model: MsgSecEventReqCase1):
        instance = MsgSecEventReq(b'test')
        mocker.patch.object(MsgSecEventReq, "_parse_msg", return_value=model.input_para)
        instance._parse_msg()
        assert model.expect == instance.msg_body

    def test_check_msg_req_para(self, model: MsgSecEventReqCase2):
        instance = MsgSecEventReq(b'test')
        instance._is_check_ok = model.input_para1
        instance.msg_body = model.input_para2
        instance._check_msg_req_para()
        assert model.expect == instance._is_check_ok

    def test_report_to_fd(self, mocker: MockerFixture, model: MsgSecEventReqCase2):
        instance = MsgSecEventReq(b'test')
        mocker.patch.object(MsgParse, "_check_log_content_valid", return_value=model.input_para1)
        instance._report_to_fd()
        assert model.expect == model.input_para2


class TestMsgSecEventRsp(TestBase):
    use_cases = {
        "test_deal_with": {
            "normal": ({"tag": 0x0000, "value": 1}, MsgErrorCode.MSG_PARA_CHECK_ERROR)
        },
        "test_packer_msg": {
            "msg_body_is_int": (False, 1),
            "msg_body_is_str": (False, "s"),
        },
    }

    MsgSecEventRspCase1 = namedtuple("MsgSecEventRspCase", "expect, input_para")

    def test_deal_with(self, model: MsgSecEventRspCase1):
        instance = MsgSecEventRsp({})
        instance.deal_with(model.input_para)
        assert model.expect == instance.msg_body

    def test_packer_msg(self, model: MsgSecEventRspCase1):
        instance = MsgSecEventRsp({})
        instance._packer_msg()
        instance.msg_body = model.input_para
        assert model.expect == instance.is_packer_ok


class TestHiSecEventProc(TestBase):
    @getLog.clear_common_log
    def test_report_hisec_event_should_failed_when_hisec_is_invalid(self):
        HiSecEventProc.report_hisec_event("1")
        ret = getLog.get_log()
        assert "Invalid HiSec event message." in ret

    @getLog.clear_common_log
    def test_report_hisec_event_should_failed_when_eventtype_is_invalid(self):
        HiSecEventProc.report_hisec_event({"eventType": 1})
        ret = getLog.get_log()
        assert "Get HiSec event type failed." in ret

    @mock.patch.object(HiSecEventProc, 'deal_hisec_message', mock.Mock(return_value=False))
    @getLog.clear_common_log
    def test_report_hisec_event_should_failed_when_eventdetail_is_invalid(self):
        HiSecEventProc.report_hisec_event({"eventType": "Rootkit attack", })
        ret = getLog.get_log()
        assert "Query event detail failed." in ret

    @mock.patch.object(HiSecEventProc, 'deal_hisec_file_privilege_escalation', mock.Mock(return_value="typeone"))
    def test_deal_hisec_message_type_one(self):
        ret = HiSecEventProc.deal_hisec_message({"eventType": "File privilege escalation", })
        assert ret == "typeone"

    @mock.patch.object(HiSecEventProc, 'deal_hisec_key_file_or_shell_file_tampering', mock.Mock(return_value="typetwo"))
    def test_deal_hisec_message_type_two(self):
        ret = HiSecEventProc.deal_hisec_message({"eventType": "Key file tampering", })
        assert ret == "typetwo"

    @mock.patch.object(HiSecEventProc, 'deal_hisec_unauthorized_root_user', mock.Mock(return_value="typethree"))
    def test_deal_hisec_message_type_three(self):
        ret = HiSecEventProc.deal_hisec_message({"eventType": "Unauthorized root user", })
        assert ret == "typethree"

    @mock.patch.object(HiSecEventProc, 'deal_hisec_rootkit_attack', mock.Mock(return_value="typefour"))
    def test_deal_hisec_message_type_four(self):
        ret = HiSecEventProc.deal_hisec_message({"eventType": "Rootkit attack", })
        assert ret == "typefour"

    @getLog.clear_common_log
    def test_deal_hisec_message_error(self):
        HiSecEventProc.deal_hisec_message({"eventType": "Root", })
        ret = getLog.get_log()
        assert "deal hisec message error" in ret

    @getLog.clear_common_log
    def test_deal_hisec_file_privilege_escalation_should_failed(self):
        HiSecEventProc.deal_hisec_file_privilege_escalation("")
        ret = getLog.get_log()
        assert "deal hisec file privilege escalation error: invalid message format" in ret

    def test_deal_hisec_file_privilege_escalation_should_ok(self):
        ret = HiSecEventProc.deal_hisec_file_privilege_escalation({"eventName": "name"})
        assert ret == "[MINOR] name, the method is None, the file is None."

    def test_deal_hisec_key_file_or_shell_file_tampering_should_failed_when_evidence_not_exists(self):
        ret = HiSecEventProc.deal_hisec_key_file_or_shell_file_tampering({})
        assert ret == "[MINOR] None, file is None."

    @getLog.clear_common_log
    def test_deal_hisec_key_file_or_shell_file_tampering_should_failed_when_evidence_is_invalid(self):
        HiSecEventProc.deal_hisec_key_file_or_shell_file_tampering({"evidence": "123"})
        ret = getLog.get_log()
        assert "deal hisec key file or shell file tampering error: message invalid format" in ret

    @getLog.clear_common_log
    def test_deal_hisec_key_file_or_shell_file_tampering_should_failed_when_attribute_is_invalid(self):
        HiSecEventProc.deal_hisec_key_file_or_shell_file_tampering({"evidence": {
            "attribute": [0], }, })
        ret = getLog.get_log()
        assert "deal hisec key file or shell file tampering error: invalid attribute" in ret

    @getLog.clear_common_log
    def test_deal_hisec_key_file_or_shell_file_tampering_should_ok(self):
        ret = HiSecEventProc.deal_hisec_key_file_or_shell_file_tampering({"evidence": {
            "attribute": [{"from": "f", }], "associatedPath": "aspath"}, })
        assert ret == "[MINOR] None, None has been moved to aspath."

    def test_deal_hisec_unauthorized_root_user(self):
        ret = HiSecEventProc.deal_hisec_unauthorized_root_user({"eventName": "evname"})
        assert ret == "[MINOR] evname, the unauthorized root user is None."

    def test_deal_hisec_rootkit_attack_when_evidence_not_exists(self):
        ret = HiSecEventProc.deal_hisec_rootkit_attack({})
        assert ret == "[MINOR] None, the rootkit name is None."

    @mock.patch.object(HiSecEventProc, "_deal_msg_rootkit_attack", mock.Mock(return_value=True))
    def test_deal_hisec_rootkit_attack_when_evidence_exists(self):
        ret = HiSecEventProc.deal_hisec_rootkit_attack({"evidence": {"feature": "123"}})
        assert ret == "[MINOR] None, the rootkit name is None, the file is ."
