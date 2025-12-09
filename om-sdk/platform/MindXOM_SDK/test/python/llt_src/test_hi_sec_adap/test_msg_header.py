from collections import namedtuple

from bin.HiSecAdap.msg_header import MsgHeaderParse, MsgHeaderPacker
from conftest import TestBase


class TestMsgHeaderParse(TestBase):
    use_cases = {
        "test_get_msg_map": {
            "normal": ({"intfVersion": 0, "msg_type": 0, "msg_seq_num": 0, "msg_body_len": 0},)
        },
        "test_pasre_msg_header": {
            "normal": (0,)
        },
        "test_pasre_interface_version": {
            "byte_data_len_lt_index_plus_msg_inter_version_size": (False, b't', 1),
            "byte_data_len_gt_index_plus_msg_inter_version_size": (4, b'test1', 2),
        },
        "test_pasre_msg_type": {
            "byte_data_len_lt_index_plus_msg_inter_version_size": (False, b't', 1),
            "byte_data_len_gt_index_plus_msg_inter_version_size": (4, b'test1', 2),
        },
        "test_pasre_msg_body_len": {
            "byte_data_len_lt_index_plus_msg_inter_version_size": (False, b't', 1),
            "byte_data_len_gt_index_plus_msg_inter_version_size": (4, b'test1', 2),
        }
    }

    MsgHeaderParse0 = namedtuple("MsgHeaderParse", "expect")
    MsgHeaderParse2 = namedtuple("MsgHeaderParse", "expect, input_para1, input_para2")

    def test_get_msg_map(self, model: MsgHeaderParse0):
        instance = MsgHeaderParse(b't')
        assert model.expect == instance.get_msg_map()

    def test_pasre_msg_header(self, model: MsgHeaderParse0):
        instance = MsgHeaderParse(b't')
        instance._pasre_msg_header()
        assert model.expect == instance.intfVersion

    def test_pasre_interface_version(self, model: MsgHeaderParse2):
        instance = MsgHeaderParse(model.input_para1)
        instance._pasre_interface_version()
        if model.input_para2 == 1:
            assert model.expect == instance.is_parse_ok
        else:
            assert model.expect == instance.index

    def test_pasre_msg_type(self, model: MsgHeaderParse2):
        instance = MsgHeaderParse(model.input_para1)
        instance._pasre_msg_type()
        if model.input_para2 == 1:
            assert model.expect == instance.is_parse_ok
        else:
            assert model.expect == instance.index

    def test_pasre_msg_body_len(self, model: MsgHeaderParse2):
        instance = MsgHeaderParse(model.input_para1)
        instance._pasre_msg_body_len()
        if model.input_para2 == 1:
            assert model.expect == instance.is_parse_ok
        else:
            assert model.expect == instance.index


class TestMsgHeaderPacker(TestBase):
    use_cases = {
        "test_get_msg_bytes": {
            "normal": (b'',)
        },
        "test_packer_msg_header": {
            "normal": (False,)
        },
        "test_packer_msg_interface_version": {
            "int_version_is_int": (b'\x01\x00\x00\x00', 1, 1),
            "int_version_is_string": (False, "sss", 2),
            "int_version_is_none": (False, None, 2)
        },
        "test_packer_msg_type": {
            "int_version_is_int": (b'\x02\x00\x00\x00', 2, 1),
            "int_version_is_string": (False, "S", 2),
            "int_version_is_none": (False, None, 2)
        },
        "test_packer_msg_seq_num": {
            "int_version_is_int": (b'\x03\x00\x00\x00', 3, 1),
            "int_version_is_string": (False, "T", 2),
            "int_version_is_none": (False, None, 2)
        },
        "test_packer_msg_body_len": {
            "int_version_is_int": (b'\x00\x00\x00\x00', 0, 1),
            "int_version_is_string": (False, "t", 2),
            "int_version_is_none": (False, None, 2)
        }
    }

    MsgHeaderPacker0 = namedtuple("MsgHeaderPacker", "expect")
    MsgHeaderPacker1 = namedtuple("MsgHeaderPacker", "expect, input_para1")
    MsgHeaderPacker2 = namedtuple("MsgHeaderPacker", "expect, input_para1, input_para2")

    def test_get_msg_bytes(self, model: MsgHeaderPacker0):
        instance = MsgHeaderPacker({})
        assert model.expect == instance.get_msg_bytes()

    def test_packer_msg_header(self, model: MsgHeaderPacker0):
        instance = MsgHeaderPacker({})
        instance._packer_msg_header()
        assert model.expect == instance.is_packer_ok

    def test_packer_msg_interface_version(self, model: MsgHeaderPacker2):
        instance = MsgHeaderPacker({"intfVersion": model.input_para1})
        instance._packer_msg_interface_version()
        if model.input_para2 == 1:
            assert model.expect == instance.msg_bytes
        else:
            assert model.expect == instance.is_packer_ok

    def test_packer_msg_type(self, model: MsgHeaderPacker2):
        instance = MsgHeaderPacker({})
        instance.msg_type = model.input_para1
        instance._packer_msg_type()
        if model.input_para2 == 1:
            assert model.expect == instance.msg_bytes
        else:
            assert model.expect == instance.is_packer_ok

    def test_packer_msg_seq_num(self, model: MsgHeaderPacker2):
        instance = MsgHeaderPacker({"msg_seq_num": model.input_para1})
        instance._packer_msg_seq_num()
        if model.input_para2 == 1:
            assert model.expect == instance.msg_bytes
        else:
            assert model.expect == instance.is_packer_ok

    def test_packer_msg_body_len(self, model: MsgHeaderPacker2):
        instance = MsgHeaderPacker({})
        instance.msg_body_len = model.input_para1
        instance._packer_msg_body_len()
        if model.input_para2 == 1:
            assert model.expect == instance.msg_bytes
        else:
            assert model.expect == instance.is_packer_ok
