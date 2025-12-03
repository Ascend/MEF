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

import pytest
from pytest_mock import MockerFixture

from common.utils.app_common_method import AppCommonMethod
from common.checkers.base_checker.abc_checker import AttrCheckerBase
from common.utils.result_base import Result
from net_manager.checkers.external_params_checker import NetManageConfigChecker
from net_manager.common_methods import ApiHelper, Serializer
from net_manager.exception import ValidateParamsError


class TestUtils:
    url = "0/v1/NetManager"
    method = "get"

    @staticmethod
    def decode():
        return '{"is_student": true}'

    @staticmethod
    def _update_general_error(a_a):
        return True


class TestUtils1:
    method = "PATCH"
    headers = {'X-Real-Ip': "1"}

    @staticmethod
    def decode():
        return '{"ManagerType": "Web"}'


class TestGetJsonInfoObj:
    CheckServerNameForbiddenCase = namedtuple("CheckServerNameForbiddenCase",
                                              "expected, server_name, get_etc_host_content_lines")
    CheckUpdateGeneralErrorCase = namedtuple("CheckUpdateGeneralErrorCase", "ret_dict")
    use_cases = {
        "test_check_server_name_forbidden": {
            "first": (True, "localhost", [Result(False), ]),
            "second": (True, "localhost1", [Result(False), ]),
            "third": (False, "localhost1", [Result(True, data=(" a  ", "::1")), ]),
        },
        "test_update_general_error": {
            "first": ({"status": 400, "message": [110206, "test"]}, ),
            "second": ({"status": 400, "message": "test"}, ),
        },
    }

    def test_get_api_operation(self):
        assert ApiHelper.get_api_operation(TestUtils)

    def test_validate_params_raise(self):
        with pytest.raises(ValidateParamsError):
            ApiHelper.validate_params(NetManageConfigChecker(), TestUtils)

    def test_validate_params_first_raise(self, mocker: MockerFixture):
        mocker.patch.object(AttrCheckerBase, "check").side_effect = Exception()
        with pytest.raises(ValidateParamsError):
            ApiHelper.validate_params(NetManageConfigChecker(), TestUtils)

    def test_validate_params(self):
        assert ApiHelper.validate_params(NetManageConfigChecker(), TestUtils1)

    def test_check_server_name_forbidden(self, mocker: MockerFixture, model: CheckServerNameForbiddenCase):
        mocker.patch.object(AppCommonMethod, "get_etc_host_content_lines").side_effect = \
            model.get_etc_host_content_lines
        ret = ApiHelper.check_server_name_forbidden(model.server_name)
        assert ret == model.expected

    def test_send_operational_log(self):
        assert not ApiHelper.send_operational_log(TestUtils1, "")

    def test_make_200_response(self):
        class TestUtils2:
            class service:

                @staticmethod
                def get_resource():
                    return '{"is_student": true}'
        with pytest.raises(RuntimeError):
            Serializer.make_200_response(TestUtils2, {"keys": 1}, 200)

    def test_make_206_response(self):
        assert Serializer.make_206_response(TestUtils, "")

    def test_make_400_response(self):
        assert Serializer.make_400_response(TestUtils, "a")

    def test_make_500_response(self):
        assert Serializer.make_500_response(TestUtils)

    def test_update_general_error(self, model: CheckUpdateGeneralErrorCase):
        with pytest.raises(RuntimeError):
            Serializer._update_general_error(Serializer, model.ret_dict)
