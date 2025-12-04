#!/usr/bin/python3
# -*- coding: UTF-8 -*-
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

from pytest_mock import MockerFixture

from bin.environ import Env
from common.checkers import IPChecker, CheckResult
from lib.Linux.systems.security_service.security_load import SecurityLoad
from ut_utils.mock_utils import mock_cdll

with mock_cdll():
    from lib.Linux.EdgeSystem.user_operate import UserOperate

env = Env()

CheckSecurityConfig = namedtuple("CheckSecurityConfig", "expect, param_dict, check_session")
PostRequest = namedtuple("PostRequest", "expect, param_dict, check_weak, check_security, set_os_pwd")


class TestMcuInfo:
    use_cases = {
        "test_check_security_config": {
            "not sub list": ([400, "Parameter is invalid."], {"a": 1}, None),
            "check ip failed": ([400, "real_ip"], {"real_ip": "192.168.2.1"}, False),
            "normal": ([200, "check success"], {"real_ip": "192.168.2.1"}, True),
        },
        "test_post_request": {
            "operate type wrong": ([400, "Parameter is invalid."], {}, None, None, None),
            "check weak": ([200, "OK"], {"oper_type": "check_weak_dict"}, [200, "OK"], None, None),
            "check_security_config": ([200, "OK"], {"oper_type": "check_security_config"}, None, [200, "OK"], None),
            "set_os_pwd": ([200, "OK"], {"oper_type": "reset_password"}, None, None, [200, "OK"]),
            "else": ([400, "operate type is invalid."], {"oper_type": "else"}, [200, "OK"], None, None),
        }
    }

    @staticmethod
    def test_check_security_config(mocker: MockerFixture, model: CheckSecurityConfig):
        mocker.patch.object(IPChecker, "check", return_value=CheckResult.make_success())
        mocker.patch.object(SecurityLoad, "check_session_request_usr_data", return_value=model.check_session)
        ret = UserOperate.check_security_config(model.param_dict)
        assert ret == model.expect

    @staticmethod
    def test_post_request(mocker: MockerFixture, model: PostRequest):
        env.__dict__.update(start_from_m2=False)
        mocker.patch.object(UserOperate, "check_weak_dict", return_value=model.check_weak)
        mocker.patch.object(UserOperate, "check_security_config", return_value=model.check_security)
        mocker.patch.object(UserOperate, "set_os_pwd", return_value=model.set_os_pwd)
        op = UserOperate()
        ret = op.post_request(model.param_dict)
        assert ret == model.expect
