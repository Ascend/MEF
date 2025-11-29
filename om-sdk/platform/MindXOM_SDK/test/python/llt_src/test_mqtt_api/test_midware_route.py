# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from fd_msg_process.midware_route import MidwareRoute


class TestGetJsonInfoObj:
    def test_add_url_rule(self):
        assert not MidwareRoute.add_url_rule("", "", True)

    def test_route(self):
        assert MidwareRoute.route("", True)
