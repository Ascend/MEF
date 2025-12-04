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
from unittest.mock import patch

import pytest
from flask.testing import FlaskClient
from pytest_mock import MockerFixture

from om_system_service.default_config import DefaultConfig
from om_system_service.default_config import RestoreConfigError
from lib_restful_adapter import LibRESTfulAdapter
from ut_utils.models import MockPrivilegeAuth
from test_bp_api.create_client import get_client

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from system_service.systems_blueprint import system_bp


CheckParam = namedtuple("CheckParam", "request_dict")
RestoreMonitorConfig = namedtuple("RestoreMonitorConfig", "request_dict, return_value")
RestoreMefConfig = namedtuple("RestoreMefConfig", "request_dict, return_value")
RestoreNetworkConfig = namedtuple("RestoreNetworkConfig", "request_dict, return_value")
RebootSystem = namedtuple("RebootSystem", "return_value")


class TestDefaultConfig:
    client: FlaskClient = get_client(system_bp)
    use_cases = {
        "test_check_param": {
            "invalid_type_of_param": ("132", ),
            "invalid_param": (
                {
                    "ReserveIP": None,
                    "Password": "",
                }, ),
        },
        "test_restore_monitor_config_failed": {
            "failed": (
                dict(),
                {"status": 400, "message": "restore config failed"},
            ),
        },
        "test_restore_mef_config_failed": {
            "busy": (
                dict(),
                {"status": 400, "message": "mef is busy"},
            ),
            "failed": (
                dict(),
                {"status": 400, "message": "restore mef config failed"},
            ),
        },
        "test_restore_network_config_failed": {
            "failed": (
                dict(),
                {"status": 400, "message": "restore network config failed"},
            ),
        },
        "test_reboot_system_failed": {
            "failed": (
                {"status": 400, "message": "reboot failed"},
            ),
        },
    }

    @staticmethod
    def test_check_param(model: CheckParam):
        with pytest.raises(RestoreConfigError) as exception_info:
            DefaultConfig.check_param(model.request_dict)
            assert "param is invalid" in str(exception_info.value)

    @staticmethod
    def test_restore_monitor_config_failed(mocker: MockerFixture, model: RestoreMonitorConfig):
        with pytest.raises(RestoreConfigError) as exception_info:
            mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
            DefaultConfig.restore_monitor_config(model.request_dict)
            assert "restore config failed" in str(exception_info.value)

    @staticmethod
    def test_restore_mef_config_failed(mocker: MockerFixture, model: RestoreMefConfig):
        with pytest.raises(RestoreConfigError) as exception_info:
            mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
            DefaultConfig.restore_mef_config(model.request_dict)
            assert model.return_value.get("message") in str(exception_info.value)

    @staticmethod
    def test_restore_network_config_failed(mocker: MockerFixture, model: RestoreNetworkConfig):
        with pytest.raises(RestoreConfigError) as exception_info:
            mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
            DefaultConfig.restore_network_config(model.request_dict)
            assert model.return_value.get("message") in str(exception_info.value)

    @staticmethod
    def test_reboot_system_failed(mocker: MockerFixture, model: RebootSystem):
        with pytest.raises(RestoreConfigError) as exception_info:
            mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
            DefaultConfig.reboot_system()
            assert model.return_value.get("message") in str(exception_info.value)

