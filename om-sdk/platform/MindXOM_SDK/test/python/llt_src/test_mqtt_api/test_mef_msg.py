# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

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
