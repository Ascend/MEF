from collections import namedtuple
import random
from pytest_mock import MockerFixture

from bin.HiSecAdap.msg import MsgParse, MsgPacker
from conftest import TestBase


class TestMsgParse(TestBase):
    use_cases = {
        "test_get_msg_map": {
            "normal": ({"intfVersion": 0, "msg_type": 0, "msg_seq_num": 0, "msg_body_len": 0, "msg_body": []},)
        },
        "test_check_log_content_valid": {
            "log_content_is_valid": (True, "111"),
            "log_content_gt_max_str_val_len": (False, "1" * 1024 * 1025),
            "log_content_contains_invalid_character": (False, random.choice(["\r", "\n", "\b", "\f", "\v", "\u007F"])),
        },
        "test_parse_tlv_list": {
            "tlv_list_len_gt_max_len_val": ([], 1, 1024 * 1024 + 1, False),
            "tlv_list_len_lt_0": ([], 1, -1, False),
            "is_parse_ok_is_false": ([], 1, 2, False),
            "normal": ([{"tag": "aaa", "length": 10, "value": 10}], 1, 14, True),
            "sub_tlv_index_ne_tlv_list_start_index_plus_tlv_list_len": ([], 1, 13, True)
        },
        "test_parse_tlv_double_list": {
            "tlv_list_len_gt_max_len_val": ([], 0, 1024 * 1024 + 1, False),
            "is_parse_ok_is_false": ([], 1, 13, False),
            "sub_tlv_index_ne_tlv_list_start_index_plus_tlv_list_len": ([], 1, 13, True),
            "normal": ([{"tag": "aaa", "length": 10, "value": []}], 1, 14, True)
        },
        "test_parse_tlv": {
            "normal": ({"tag": "aaa", "length": 10, "value": 10}, 0)
        },
        "test_parse_tag": {
            "end_idx_gt_byte_data_len": (-1, b'test', 3),
            "ret_gt_0x008e": (-1, b'10001111', 0),
            "ret_lt_0": (-1, b'101', 0)
        },
        "test_parse_length": {
            "end_idx_gt_byte_data_len": (-1, 3)
        },
        "test_parse_value": {
            "tag_in_int_value_and_end_idx_gt_tlv_value_index_plus_value_len": (-1, 5, 5, 0x0071),
            "tag_in_string_value_and_end_idx_gt_tlv_value_index_plus_value_len": ("", 5, 5, 0x0074),
            "tag_not_in_string_value_and_not_in_int_value": (0, 5, 5, 11)
        }
    }

    MsgParseCase0 = namedtuple("MsgParseCase", "expect")
    MsgParseCase1 = namedtuple("MsgParseCase", "expect, input_para")
    MsgParseCase2 = namedtuple("MsgParseCase", "expect, input_para1, input_para2")
    MsgParseCase3 = namedtuple("MsgParseCase", "expect, input_para1, input_para2, input_para3")

    def test_get_msg_map(self, model: MsgParseCase0):
        instance = MsgParse(b'test')
        assert model.expect == instance.get_msg_map()

    def test_check_log_content_valid(self, model: MsgParseCase1):
        instance = MsgParse(b'test')
        assert model.expect == instance._check_log_content_valid(model.input_para)

    def test_parse_tlv_list(self, mocker: MockerFixture, model: MsgParseCase3):
        mocker.patch.object(MsgParse, "_parse_tag", return_value="aaa")
        mocker.patch.object(MsgParse, "_parse_length", return_value=10)
        mocker.patch.object(MsgParse, "_parse_value", return_value=10)
        instance = MsgParse(b'test')
        instance.is_parse_ok = model.input_para3
        assert model.expect == instance._parse_tlv_list(model.input_para1, model.input_para2)

    def test_parse_tlv_double_list(self, mocker: MockerFixture, model: MsgParseCase3):
        mocker.patch.object(MsgParse, "_parse_tag", return_value="aaa")
        mocker.patch.object(MsgParse, "_parse_length", return_value=10)
        mocker.patch.object(MsgParse, "_parse_tlv_list", return_value=[])
        instance = MsgParse(b'test')
        instance.is_parse_ok = model.input_para3
        assert model.expect == instance._parse_tlv_double_list(model.input_para1, model.input_para2)

    def test_parse_tlv(self, mocker: MockerFixture, model: MsgParseCase1):
        mocker.patch.object(MsgParse, "_parse_tag", return_value="aaa")
        mocker.patch.object(MsgParse, "_parse_length", return_value=10)
        mocker.patch.object(MsgParse, "_parse_value", return_value=10)
        instance = MsgParse(b'test')
        assert model.expect == instance._parse_tlv(model.input_para)

    def test_parse_tag(self, model: MsgParseCase2):
        instance = MsgParse(b'test')
        instance.byte_data_len = 4
        instance.byte_data = model.input_para1
        assert model.expect == instance._parse_tag(model.input_para2)

    def test_parse_length(self, model: MsgParseCase1):
        instance = MsgParse(b'test')
        instance.byte_data_len = 4
        assert model.expect == instance._parse_length(model.input_para)

    def test_parse_value(self, model: MsgParseCase3):
        instance = MsgParse(b'test')
        instance.byte_data_len = 8
        assert model.expect == instance._parse_value(model.input_para1, model.input_para2, model.input_para3)


class TestMsgPacker(TestBase):
    use_cases = {
        "test_packer_list_map": {
            "normal": (b'aaa', "1")
        },
        "test_packer_map": {
            "tag_is_none": (b'', {}),
            "value_is_none": (b'', {"tag": 1}),
            "tag_and_value_is_not_none": (b'\x00\x01\x00\x04\x01\x00\x00\x00', {"tag": 1, "value": 1})
        },
        "test_packer_tag": {
            "tag_is_int_1": (b'\x00\x01', 1),
            "tag_is_str_1": (b'\x00\x01', "1"),
            "tag_is_str_s": (0, "s")
        },
        "test_packer_value": {
            "value_is_list": (b'aaa', [1, 2]),
            "value_is_dict": (b'aaa', {}),
            "value_is_int": (b'o\x00\x00\x00', 111),
            "value_is_str": (b'111', "111"),
            "value_is_tuple": (b'', (1,))
        },
        "test_get_msg_bytes": {
            "normal": (b'test', b'test')
        },
        "test_assemble_msg_header_and_body": {
            "normal": (b'\x00\x00\x00\x00',)
        },
        "test_packer_msg_body": {
            "msg_body_is_list": (b'\x00\x00', [], 1),
            "msg_body_is_dict": (b'\x00\x00', {}, 1),
            "msg_body_is_not_dict_and_not_list": (False, 'test', 2)
        }
    }

    MsgPackerCase0 = namedtuple("MsgPackerCase", "expect")
    MsgPackerCase1 = namedtuple("MsgPackerCase", "expect, input_para")
    MsgPackerCase2 = namedtuple("MsgPackerCase", "expect, input_para1, input_para2")

    def test_packer_list_map(self, mocker: MockerFixture, model: MsgPackerCase1):
        mocker.patch.object(MsgPacker, "packer_map", return_value=b"aaa")
        instance = MsgPacker({})
        assert model.expect == instance.packer_list_map(model.input_para)

    def test_packer_map(self, model: MsgPackerCase1):
        instance = MsgPacker({})
        assert model.expect == instance.packer_map(model.input_para)

    def test_packer_tag(self, model: MsgPackerCase1):
        instance = MsgPacker({})
        assert model.expect == instance.packer_tag(model.input_para)

    def test_packer_value(self, mocker: MockerFixture, model: MsgPackerCase1):
        mocker.patch.object(MsgPacker, "packer_list_map", return_value=b"aaa")
        mocker.patch.object(MsgPacker, "packer_map", return_value=b"aaa")
        instance = MsgPacker({})
        assert model.expect == instance.packer_value(model.input_para)

    def test_get_msg_bytes(self, model: MsgPackerCase1):
        instance = MsgPacker({})
        instance.msg_bytes = model.input_para
        assert model.expect == instance.get_msg_bytes()

    def test_assemble_msg_header_and_body(self, model: MsgPackerCase0):
        instance = MsgPacker({})
        instance._assemble_msg_header_and_body()
        assert model.expect == instance.msg_bytes

    def test_packer_msg_body(self, mocker: MockerFixture, model: MsgPackerCase2):
        mocker.patch.object(MsgPacker, "packer_list_map", return_value=b'\x00\x00')
        mocker.patch.object(MsgPacker, "packer_map", return_value=b'\x00\x00')
        instance = MsgPacker({})
        instance.msg_body = model.input_para1
        instance._packer_msg_body()
        if model.input_para2 == 1:
            assert model.expect == instance.msg_body_bytes
        else:
            assert model.expect == instance.is_packer_ok
