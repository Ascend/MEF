# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from pytest_mock import MockerFixture

from fd_msg_process.fd_add_route import add_midware_route
from fd_msg_process.midware_urls import MidwareUris


class TestGetJsonInfoObj:
    def test_publish_ws_msg(self, mocker: MockerFixture):
        mocker.patch.object(MidwareUris, "mid_ware_add_route", side_effect=[True, ])
        assert not add_midware_route()
