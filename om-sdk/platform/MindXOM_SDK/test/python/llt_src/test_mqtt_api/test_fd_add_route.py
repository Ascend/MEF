# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from pytest_mock import MockerFixture

from fd_msg_process.fd_add_route import add_midware_route
from fd_msg_process.midware_urls import MidwareUris


class TestGetJsonInfoObj:
    def test_publish_ws_msg(self, mocker: MockerFixture):
        mocker.patch.object(MidwareUris, "mid_ware_add_route", side_effect=[True, ])
        assert not add_midware_route()
