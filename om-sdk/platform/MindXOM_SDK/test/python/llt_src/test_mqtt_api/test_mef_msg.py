# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from collections import namedtuple

from mef_msg_process.mef_msg import MefMsgData


class TestUtils:
    @staticmethod
    def to_dict():
        return "1"


class TestGetJsonInfoObj:
    def test_gen_necessary_files(self):
        CopyFile = namedtuple("CopyFile1", ["header", "route", "content"])
        backup_restore_base_two = CopyFile(*(TestUtils, TestUtils, {"header": "1"}))
        assert MefMsgData.to_ws_msg_str(backup_restore_base_two) == \
               '{"header": "1", "route": "1", "content": "{\\"header\\": \\"1\\"}"}'
