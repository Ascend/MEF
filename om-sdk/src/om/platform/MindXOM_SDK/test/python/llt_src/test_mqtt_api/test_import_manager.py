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
from unittest.mock import patch

import mock
import pytest
from pytest_mock import MockerFixture

import cert_manager.cert_clib_mgr as mgr
from net_manager.constants import NetManagerConstants
from net_manager.exception import DataCheckException
from net_manager.manager.import_manager import CertImportManager, ImportManagerBase
from net_manager.manager.net_cfg_manager import NetCfgManager
from test_table_data_checker import TestUtils1 as TestUtils11


class TestUtils6(CertImportManager, ):
    source = ""
    contents = "a"


class TestUtils7(TestUtils6):
    @staticmethod
    def cert_getcertinfo(*args):
        return 0
    source = NetManagerConstants.WEB
    contents_name = "cert_contents"
    get_extend_certinfo = cert_getcertinfo


class TestGetJsonInfoObj:
    CheckUpdateContentsCase = namedtuple("CheckUpdateContentsCase",
                                         "tested_class, param_key, input_param, excepted_success")

    def test_import_deal(self):
        assert not ImportManagerBase.import_deal(ImportManagerBase)

    def test_update_contents(self):
        assert not ImportManagerBase.update_contents(ImportManagerBase)

    def test_check_contents_ex(self):
        with pytest.raises(Exception):
            assert not ImportManagerBase.check_contents(CertImportManager)

    def test_check_contents(self):
        with pytest.raises(DataCheckException):
            assert not ImportManagerBase.check_contents(TestUtils6)

    def test_check_status_is_ready(self, mocker: MockerFixture):
        mocker.patch.object(NetCfgManager, "get_net_cfg_info", return_value=TestUtils11)
        assert CertImportManager.check_status_is_ready()

    @patch.object(mgr, "CDLL", mock.Mock(return_value=TestUtils7))
    def test_update_contents(self, mocker: MockerFixture):
        mocker.patch.object(CertImportManager, "check_multi_cert", return_value=True)
        assert not CertImportManager.check_contents(TestUtils7)
        with pytest.raises(Exception):
            assert not CertImportManager.check_contents(TestUtils6)
