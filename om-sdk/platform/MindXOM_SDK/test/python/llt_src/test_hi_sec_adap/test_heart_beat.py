from collections import namedtuple

from bin.HiSecAdap.constant import MsgErrorCode
from bin.HiSecAdap.heart_beat import MsgHeartBeatReq, MsgHeartBeatRsp
from conftest import TestBase


class TestMsgHeartBeatReq(TestBase):
    use_cases = {
        "test_deal_with": {
            "is_parse_ok_is_true": (MsgErrorCode.MSG_PARSE_OR_PACKER_OK, True),
            "is_parse_ok_is_false": (MsgErrorCode.MSG_PARSE_ERROR, False)
        },
        "test_parse_msg": {
            "is_parse_ok_is_false": (None, False),
            "is_parse_ok_is_True": (None, True)
        }
    }

    MsgHeartBeatReqCase = namedtuple("MsgHeartBeatReqCase", "expect, is_parse_ok_value")

    def test_deal_with(self, model: MsgHeartBeatReqCase):
        instance = MsgHeartBeatReq(b'ut')
        instance.is_parse_ok = model.is_parse_ok_value
        assert model.expect == instance.deal_with()

    def test_parse_msg(self, model: MsgHeartBeatReqCase):
        instance = MsgHeartBeatReq(b'ut')
        instance.is_parse_ok = model.is_parse_ok_value
        assert model.expect == instance._parse_msg()


class TestMsgHeartBeatRsp(TestBase):
    use_cases = {
        "test_deal_with": {
            "ret_code_value_is_0": (0x00000002, MsgErrorCode.MSG_PARSE_OR_PACKER_OK),
            "ret_code_value_is_not_0": (0, MsgErrorCode.MSG_PARSE_ERROR)
        },
        "test_parse_msg": {
            "ret_code_value_is_0": (None, False),
            "ret_code_value_is_not_0": (None, True)
        }
    }

    MsgHeartBeatRspCase = namedtuple("MsgHeartBeatRspCase", "expect, ret_code_value")

    def test_deal_with(self, model: MsgHeartBeatRspCase):
        instance = MsgHeartBeatRsp({"intfVersion": 0})
        instance.deal_with(model.ret_code_value)
        assert model.expect == instance.msg_type

    def test_parse_msg(self, model: MsgHeartBeatRspCase):
        instance = MsgHeartBeatRsp({"intfVersion": 0})
        instance.deal_with(model.ret_code_value)
        assert model.expect == instance._packer_msg()
