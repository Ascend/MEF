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

from pytest_mock import MockerFixture

from common.kmc_lib.tlsconfig import TlsConfig
from common.utils.result_base import Result
from net_manager.manager.fd_cert_manager import FdCertManager


class TestUtils4:
    cert_contents = "123"
    crl_contents = ""


class TestUtils3:
    @staticmethod
    def _gen_ssl_context(a_a, b_b):
        raise Exception()

    @staticmethod
    def get_current_using_cert():
        return TestUtils4


class TestUtils2:
    class cert_manager:
        @staticmethod
        def get_obj_by_source(a_a):
            return

        @staticmethod
        def get_all():
            return [TestUtils1(), ]

    @staticmethod
    def get_current_using_cert():
        return

    @staticmethod
    def _gen_ssl_context(a_a, b_b):
        return True


class TestUtils1:
    id = "0"
    cert_contents = "123"
    crl_contents = "123"

    @staticmethod
    def get_current_using_cert():
        return

    class cert_manager:
        @staticmethod
        def get_obj_by_source(a_a):
            return [TestUtils1, ]


class TestUtils:
    @staticmethod
    def get_current_using_cert():
        return TestUtils1

    @staticmethod
    def is_socket_connect_pass(a_a):
        return True

    class cert_manager:
        @staticmethod
        def get_obj_by_source(a_a):
            return "12345"

        @staticmethod
        def get_all():
            return [TestUtils1(), ]


class TestGetJsonInfoObj:
    CheckGetCertInfoCase = namedtuple("CheckGetCertInfoCase", "class1")
    CheckGetClientSslContextCase = namedtuple("CheckGetClientSslContextCase", "expected, class1")
    CheckVerifyCrlAgainstCaCase = namedtuple("CheckVerifyCrlAgainstCaCase", "class1, source")
    CheckGenSslContextCase = namedtuple("CheckGenSslContextCase", "class1")
    use_cases = {
        "test_get_cert_info": {
            "first": (TestUtils, ),
            "all": (TestUtils1, ),
        },
        "test_get_client_ssl_context": {
            "True": (True, TestUtils2,),
            "False": (False, TestUtils3,),
        },
        "test_verify_crl_against_ca": {
            "first": (TestUtils1, "Web"),
            "second": (TestUtils2, "Web"),
            "third": (TestUtils, "Web"),
            "fifth": (TestUtils, "FusionDirector"),
        },
        "test_gen_ssl_context": {
            "final": (TestUtils,),
            "second": (TestUtils3,),
            "first": (TestUtils1,),
        },
    }

    def test_is_socket_connect_pass(self):
        CopyFile = namedtuple("CopyFile1", ["addr", "SOCKET_CONNECT_TIMEOUT"])
        backup_restore_base_two = CopyFile(*(["127.0.1.1", 8080], 5))
        assert not FdCertManager.is_socket_connect_pass(backup_restore_base_two, "")

    def test_is_cert_pool_full(self):
        ret = FdCertManager.is_cert_pool_full(TestUtils)
        assert ret

    def test_get_current_using_cert(self, mocker: MockerFixture):
        mocker.patch.object(TlsConfig, "get_client_context_with_cadata", return_value=[True, Result(False)])
        assert FdCertManager.get_current_using_cert(TestUtils)

    def test_get_cert_info(self):
        assert not FdCertManager.get_cert_info(TestUtils2, "FusionDirector")

    def test_get_client_ssl_context(self, model: CheckGetClientSslContextCase):
        ret = FdCertManager.get_client_ssl_context(model.class1)
        assert model.expected == bool(ret)

    def test_verify_crl_against_ca(self, model: CheckVerifyCrlAgainstCaCase):
        assert not FdCertManager.verify_crl_against_ca(model.class1, "", model.source)

    def test_gen_ssl_context(self, mocker: MockerFixture, model: CheckGenSslContextCase):
        mocker.patch("os.open")
        mocker.patch("os.fdopen").return_value.__enter__.return_value.write.side_effect = "abc"
        assert not FdCertManager._gen_ssl_context(model.class1, "test.crt", "test.crl")
