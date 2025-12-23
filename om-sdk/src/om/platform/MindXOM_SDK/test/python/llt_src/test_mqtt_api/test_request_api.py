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

from collections import namedtuple

import pytest
from pytest_mock import MockerFixture

from common.file_utils import FileCreate
from net_manager.manager.net_cfg_manager import NetCfgManager
from net_manager.request_api import set_response_operational_log, rf_config_net_manage, rf_get_net_manage_config, \
    rf_get_node_id, rf_get_fd_cert, rf_import_fd_cert, rf_import_fd_crl


class TestUtils:
    node_id = "1"
    ip = "1"
    port = "1"
    filename = "a"

    @staticmethod
    def to_dict_for_query():
        return {"a": "b"}


class TestGetJsonInfoObj:
    CheckRfConfigNetManageCase = namedtuple("CheckRfImportFdCertCase", "locked")
    CheckRfImportFdCertCase = namedtuple("CheckRfImportFdCertCase", "locked, create_dir")
    CheckRfImportFdCrlCase = namedtuple("CheckRfImportFdCrlCase", "locked, create_dir")
    use_cases = {
        "test_rf_config_net_manage": {
            "first": (True, ),
            "second": (False, ),
        },
        "test_rf_import_fd_cert": {
            "first": (True, False),
            "second": (False, False),
            "third": (False, True),
        },
        "test_rf_import_fd_crl": {
            "first": (True, False),
            "second": (False, False),
            "third": (False, True),
        },
    }

    def test_set_response_operational_log(self):
        with pytest.raises(RuntimeError):
            set_response_operational_log("")

    def test_rf_config_net_manage(self, mocker: MockerFixture, model: CheckRfConfigNetManageCase):
        mocker.patch("net_manager.request_api.net_manager_lock").locked.return_value = model.locked
        with pytest.raises(RuntimeError):
            rf_config_net_manage()

    def test_rf_get_net_manage_config(self, mocker: MockerFixture):
        mocker.patch.object(NetCfgManager, "get_net_cfg_info").side_effect = TestUtils
        with pytest.raises(RuntimeError):
            rf_get_net_manage_config()

    def test_rf_get_node_id(self, mocker: MockerFixture):
        mocker.patch.object(NetCfgManager, "get_net_cfg_info").side_effect = TestUtils
        with pytest.raises(RuntimeError):
            rf_get_node_id()

    def test_rf_get_fd_cert(self, mocker: MockerFixture):
        mocker.patch.object(NetCfgManager, "get_net_cfg_info").side_effect = TestUtils
        with pytest.raises(RuntimeError):
            rf_get_fd_cert()

    def test_rf_import_fd_cert(self, mocker: MockerFixture, model: CheckRfImportFdCertCase):
        mocker.patch("net_manager.request_api.import_cert_lock").locked.return_value = model.locked
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        with pytest.raises(RuntimeError):
            rf_import_fd_cert()

    def test_rf_import_fd_crl(self, mocker: MockerFixture, model: CheckRfImportFdCrlCase):
        mocker.patch("net_manager.request_api.import_crl_lock").locked.return_value = model.locked
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        with pytest.raises(RuntimeError):
            rf_import_fd_crl()
