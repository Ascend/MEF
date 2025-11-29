# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from pytest_mock import MockerFixture

from common.utils.exec_cmd import ExecCmd
from utils import get_login_user


class TestUtils:
    def test_get_login_user(self, mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output", return_value=[0, "1 1 1 1 1"])
        ret = get_login_user()
        assert ret == ('root', '1')
