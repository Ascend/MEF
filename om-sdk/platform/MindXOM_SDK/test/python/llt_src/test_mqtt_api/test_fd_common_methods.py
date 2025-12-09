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

from concurrent.futures.thread import ThreadPoolExecutor

from pytest_mock import MockerFixture

from fd_msg_process.fd_common_methods import FdCommonMethod, publish_ws_msg, MidwareErrCode
from net_manager.manager.fd_cfg_manager import FdCfgManager


class TestUtils:
    @staticmethod
    def topic():
        return False


class TestGetJsonInfoObj:
    def test_run(self, mocker: MockerFixture):
        mocker.patch.object(FdCfgManager, "get_cur_fd_server_name").side_effect = [True, ]
        assert not FdCommonMethod.contains_forbidden_domain(("", ))

    def test_publish_ws_msg(self, mocker: MockerFixture):
        mocker.patch.object(ThreadPoolExecutor, "submit", side_effect=Exception())
        assert not publish_ws_msg(TestUtils)

    def test_midware_generate_err_msg(self):
        assert MidwareErrCode.midware_generate_err_msg(611, "") == 'ERR.611, '
