import os
from collections import namedtuple

import pytest
from pytest_mock import MockerFixture

from bin.HiSecAdap.uds import MsgDeal, UdsServer
from bin.HiSecAdap.uds import start_uds
from conftest import TestBase
from ut_utils.mock_utils import mock_check_input_path_valid
from ut_utils.mock_utils import mock_path_exists


class TestMsgDeal(TestBase):
    use_cases = {
        "test_deal_msg": {
            "ret_0_is_false": (None, [False, 1], int),
            "ret_0_is_true": (None, [True, 1], int),
        },
        "test_get_rsp": {
            "one_msg_data_lt_16": ([False, "invalid msg"], b'11111111'),
            "msg_type_is_invalid": ([False, "msg not support now"], b'1111111111111111'),
        },
    }

    MsgDealCase1 = namedtuple("MsgDealCase", "expect, input_para1")
    MsgDealCase2 = namedtuple("MsgDealCase", "expect, input_para1, input_para2")

    def test_deal_msg(self, mocker: MockerFixture, model: MsgDealCase2):
        instance = MsgDeal()
        mocker.patch.object(MsgDeal, "_get_rsp", return_value=model.input_para1)
        assert model.expect == instance.deal_msg(model.input_para2, None)

    def test_get_rsp(self, model: MsgDealCase1):
        instance = MsgDeal()
        assert model.expect == instance._get_rsp(model.input_para1)


class TestUdsServer(TestBase):
    use_cases = {
        "test_init_server": {
            "socket_path_is_invalid": ([False, "socketpath is invalid"], 123, int),
            "func_is_invalid": ([False, "deal func is invalid"], "test", 123),
            "normal": ([True, ""], "test", int)
        },
        "test_get_one_msg_and_deal": {
            "cur_msg_buffer_len_is_gt_max_msg_buffer_len": (None, b'1' * 10 * 1024 * 1025),
            "cur_msg_body_len_is_gt_max_msg_body_len": (None, b'1' * 16)
        },
        "test_start_uds": {
            "socket_exist_but_invalid": (None, True, False),
            "socket_exist_and_valid": (None, True, True)
        },
    }

    UdsServerCase1 = namedtuple("UdsServerCase", "expect, input_para1")
    UdsServerCase2 = namedtuple("UdsServerCase", "expect, input_para1, input_para2")
    UdsServerCase3 = namedtuple("UdsServerCase", "expect, input_para1, input_para2, input_para3")
    StartUdsCase = namedtuple("StartUdsCase", "expect, exists, path_valid")

    @staticmethod
    def test_init_server(mocker: MockerFixture, model: UdsServerCase2):
        instance = UdsServer()
        mocker.patch.object(UdsServer, "_start_socket_server", return_value=None)
        assert model.expect == instance.init_server(model.input_para1, model.input_para2)

    @staticmethod
    def test_get_one_msg_and_deal(model: UdsServerCase1):
        instance = UdsServer()
        instance._msg_buffer = model.input_para1
        with pytest.raises(ValueError):
            instance._get_one_msg_and_deal()

    @staticmethod
    def test_start_uds(mocker: MockerFixture, model: StartUdsCase):
        mock_path_exists(mocker, return_value=model.exists)
        mock_check_input_path_valid(mocker, return_value=model.path_valid)
        mocker.patch.object(os, "remove")
        mocker.patch.object(UdsServer, "_start_socket_server", return_value=None)
        assert start_uds() == model.expect
