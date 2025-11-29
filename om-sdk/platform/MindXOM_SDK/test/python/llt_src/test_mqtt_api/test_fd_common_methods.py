# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

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
